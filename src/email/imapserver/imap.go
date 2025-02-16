package imapserver

import (
	"bytes"
	"errors"
	"fmt"
	resource "github.com/SongZihuan/anonymous-message"
	"github.com/SongZihuan/anonymous-message/src/database"
	"github.com/SongZihuan/anonymous-message/src/email/smtpserver"
	"github.com/SongZihuan/anonymous-message/src/flagparser"
	"github.com/SongZihuan/anonymous-message/src/messageutils"
	"github.com/SongZihuan/anonymous-message/src/reqrate"
	"github.com/SongZihuan/anonymous-message/src/sender"
	"github.com/SongZihuan/anonymous-message/src/systemnotify"
	"github.com/SongZihuan/anonymous-message/src/utils"
	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-message"
	"github.com/emersion/go-message/mail"
	"github.com/jaytaylor/html2text"
	"io"
	"mime"
	"net"
	"strings"
	"time"
)

// RateExp: 计算周期
// RateMaxCount: 周期内最大发件数量
const (
	RateExp      time.Duration = 5 * time.Minute
	RateMaxCount int64         = 3
)

const DefaultImapCycleWithIdleTime = 5 * time.Minute
const DefaultImapCycleTime = 15 * time.Second
const DefaultImapNoopTime = 5 * time.Second

const (
	StatusStop int = iota
	StatusConnectBlock
	StatusConnectError
	StatusCommandError
)

var DefaultFromAddress *mail.Address

func StartIMAPServer() (chan bool, error) {
	if flagparser.IMAPAddress == "" || flagparser.IMAPUser == "" || flagparser.IMAPRecipient == "" || flagparser.IMAPMailBox == "" {
		return make(chan bool, 1), nil
	}

	recipientList := strings.Split(flagparser.IMAPRecipient, ",")
	recList := make([]string, 0, len(recipientList))

	for _, rec := range recipientList {
		rec = strings.TrimSpace(rec)

		if !utils.IsValidEmail(rec) {
			fmt.Printf("%s is not a valid email, ignore\n", rec)
			continue
		}

		recList = append(recList, rec)
	}

	if len(recList) == 0 {
		return nil, fmt.Errorf("no any valid recipient")
	}

	DefaultFromAddress = &mail.Address{
		Name:    resource.Name,
		Address: recList[0],
	}

	stopchan := make(chan bool, 1)

	go func() {
		var commandErrorCount = 0

	MainCycle:
		for {
			status := func() int { // main cycle
				defer func() {
					if r := recover(); r != nil {
						if _err, ok := r.(error); ok {
							fmt.Printf("操作 IMAP 服务时出现致命错误: %s\n", _err.Error())
						} else {
							fmt.Printf("操作 IMAP 服务时出现致命错误（非error）: %v\n", r)
						}
					}
				}()

				unilateralData := make(chan *imapclient.UnilateralDataMailbox, 10)
				option := &imapclient.Options{
					UnilateralDataHandler: &imapclient.UnilateralDataHandler{
						Mailbox: func(data *imapclient.UnilateralDataMailbox) {
							go func() {
								if utils.IsChanOpen(unilateralData) {
									unilateralData <- data
								}
							}()
						},
					},
				}

				imapClient, err := imapclient.DialTLS(flagparser.IMAPAddress, option)
				if err != nil {
					imapClient, err = imapclient.DialStartTLS(flagparser.IMAPAddress, option)
					if err != nil {
						imapClient, err = imapclient.DialInsecure(flagparser.IMAPAddress, option)
						if err != nil {
							fmt.Printf("connect imap server failed: %s", err)
							return StatusConnectError // return main cycle
						}
					}
				}
				defer func() {
					if imapClient != nil {
						_ = imapClient.Close()
					}
				}()

				user := flagparser.IMAPUser
				password := flagparser.IMAPPassword

				if user == flagparser.SMTPUser && password == "" {
					password = flagparser.SMTPPassword
				}

				err = imapClient.Login(user, password).Wait()
				if err != nil {
					fmt.Printf("login imap server failed: %s", err.Error())
					return StatusConnectError // return main cycle
				}
				defer func() {
					if imapClient != nil {
						_ = imapClient.Logout().Wait()
					}
				}()

				caps, err := imapClient.Capability().Wait()
				if err != nil {
					fmt.Printf("capabilities: %s", err.Error())
					if isTemporaryNetError(err) {
						return StatusConnectError // return main cycle
					}
					return StatusCommandError // return main cycle
				}

				supportIMAP4rev2 := caps.Has(imap.CapIMAP4rev2)
				supportESEARCH := caps.Has(imap.CapESearch)
				supportIDLE := caps.Has(imap.CapIdle)

				if supportIMAP4rev2 {
					fmt.Println("服务端 支持：", imap.CapIMAP4rev2)
				} else {
					fmt.Println("服务端 不支持：", imap.CapIMAP4rev2)
				}

				if supportESEARCH {
					fmt.Println("服务端 支持：", imap.CapESearch)
				} else {
					fmt.Println("服务端 不支持：", imap.CapESearch)
				}

				if supportIDLE {
					fmt.Println("服务端 支持：", imap.CapIdle)
				} else {
					fmt.Println("服务端 不支持：", imap.CapIdle)
				}

				mailboxes, err := imapClient.List("", "%", nil).Collect()
				if err != nil {
					fmt.Printf("failed to list mailboxes: %s", err.Error())
					return StatusConnectError // return main cycle
				}

				MailBox := flagparser.IMAPMailBox

				hasMailBox := func() bool {
					for _, mailbox := range mailboxes {
						if mailbox.Mailbox == MailBox {
							return true
						}
					}
					return false
				}()

				if !hasMailBox {
					fmt.Printf("failed to list mailboxes: %s", err.Error())
					if isTemporaryNetError(err) {
						return StatusConnectError // return main cycle
					}
					return StatusCommandError // return main cycle
				}

				_, err = imapClient.Select(MailBox, nil).Wait()
				if err != nil {
					fmt.Printf("fail to select mail box: %s", err.Error())
					if isTemporaryNetError(err) {
						return StatusConnectError // return main cycle
					}
					return StatusCommandError // return main cycle
				}

				commandErrorCount = 0 // 清零

				for {
					func() { // msg search cycle
						defer func() {
							if r := recover(); r != nil {
								if err, ok := r.(error); ok {
									fmt.Printf("imap panic error: %s\n", err.Error())
								} else {
									fmt.Printf("imap panic: %v\n", r)
								}
								return // return msg search cycle
							}
						}()

						now := time.Now().In(flagparser.TimeZone())

						_, err = imapClient.Select(MailBox, nil).Wait()
						if err != nil {
							return // return msg search cycle
						}

						seqSet, err := imapClient.Search(&imap.SearchCriteria{
							Since:   now.Add(-1 * 24 * time.Hour),
							NotFlag: []imap.Flag{imap.FlagSeen},
						}, nil).Wait()
						if err != nil {
							fmt.Printf("search failed: %s\n", err.Error())
							return // return msg search cycle
						}

						// requires IMAP4rev2 or ESEARCH
						if (supportESEARCH || supportIMAP4rev2) && seqSet.Count <= 0 {
							return // return msg cycle
						}

						switch set := seqSet.All.(type) {
						case imap.SeqSet:
							s, ok := set.Nums()
							if !ok {
								return // return msg search cycle
							} else if len(s) == 0 {
								return
							}
						case imap.UIDSet:
							s, ok := set.Nums()
							if !ok {
								return // return msg search cycle
							} else if len(s) == 0 {
								return // return msg search cycle
							}
						default:
							return // return msg search cycle
						}

						msgCMD := imapClient.Fetch(seqSet.All, &imap.FetchOptions{
							Flags:    true,
							Envelope: true,
							BodySection: []*imap.FetchItemBodySection{
								&imap.FetchItemBodySection{}, // 获取整个正文
							},
						})

						processSeqSet := imap.SeqSetNum()
						processSeqLen := 0

					MsgCycle:
						for {
							msg := msgCMD.Next()
							if msg == nil {
								break MsgCycle
							}

							func() { // msg reaad cycle
								buf, err := msg.Collect()
								if err != nil {
									return // return msg read cycle
								}

								subject := buf.Envelope.Subject
								messageID := buf.Envelope.MessageID
								messageDate := buf.Envelope.Date.In(flagparser.TimeZone())

								if buf.Envelope.To == nil || len(buf.Envelope.To) == 0 {
									return // return msg read cycle
								}

								toAddr := func() *mail.Address { // to addr getter
									for _, to := range buf.Envelope.To {
										for _, rec := range recList {
											if to.Addr() != "" && utils.IsValidEmail(to.Addr()) && to.Addr() == rec {
												return &mail.Address{
													Name:    to.Name,
													Address: to.Addr(),
												} // return to add getter
											}
										}
									}

									return nil // return to addr getter
								}()

								if toAddr == nil {
									return // return msg read cycle
								}

								if m, _ := database.FindIMAPMessageID(messageID); m != nil {
									// 消息已经处理过
									return // return msg read cycle
								}

								if buf.Envelope.Sender == nil || len(buf.Envelope.Sender) == 0 || buf.Envelope.Sender[0].Addr() == "" || !utils.IsValidEmail(buf.Envelope.Sender[0].Addr()) {
									return // return msg read cycle
								}

								senderAddr := &mail.Address{
									Name:    buf.Envelope.Sender[0].Name,
									Address: buf.Envelope.Sender[0].Addr(),
								}

								var fromAddr *mail.Address
								if buf.Envelope.From != nil || len(buf.Envelope.From) == 0 || buf.Envelope.From[0].Addr() == "" || !utils.IsValidEmail(buf.Envelope.From[0].Addr()) {
									fromAddr = &mail.Address{
										Name:    senderAddr.Name,
										Address: senderAddr.Address,
									}
								} else {
									fromAddr = &mail.Address{
										Name:    buf.Envelope.From[0].Name,
										Address: buf.Envelope.From[0].Addr(),
									}
								}

								var replyAddr *mail.Address
								if buf.Envelope.ReplyTo != nil && len(buf.Envelope.ReplyTo) == 0 {
									replyAddr = &mail.Address{
										Name:    fromAddr.Name,
										Address: fromAddr.Address,
									}
								} else {
								ReplyToCycle:
									for _, r := range buf.Envelope.ReplyTo {
										if r.Addr() != "" && utils.IsValidEmail(r.Addr()) {
											replyAddr = &mail.Address{
												Name:    r.Name,
												Address: r.Addr(),
											}
											break ReplyToCycle
										}
									}

									if replyAddr == nil {
										replyAddr = &mail.Address{
											Name:    fromAddr.Name,
											Address: fromAddr.Address,
										}
									}
								}

								errFunc := func(errMsg string) error {
									_, err := smtpserver.SendErrorMsg(subject, messageID, fromAddr, toAddr, replyAddr, errMsg)
									if err != nil && errors.Is(err, smtpserver.ErrRateLimit) {
										return nil
									} else if err != nil {
										return err
									}

									return nil
								}

								if reqrate.CheckIMAPRate(buf.Envelope, RateExp) > RateMaxCount {
									_ = errFunc("信件发送速度过快、次数过多")
									return // return msg read cycle
								}

								var body []byte
							BodySessionCycle:
								for session, bd := range buf.BodySection {
									if session.Specifier == "" {
										body = bd
										break BodySessionCycle
									}
								}

								mailmsg, err := mail.CreateReader(bytes.NewReader(body))
								if err != nil && message.IsUnknownCharset(err) {
									_ = errFunc("邮件编码错误，我们只接受UTF-8编码")
									return // return msg read cycle
								} else if err != nil {
									_ = errFunc("无法读取邮件内容，请检查您的邮件以及其编码，我们只接受UTF-8编码")
									return // return msg read cycle
								}
								defer func() {
									_ = mailmsg.Close()
								}()

								var contentType string
								var mimeType string
								var mimeParams map[string]string
								var encoding string
								var bodyStr string
								var bodySafe bool

							BodyPartCycle:
								for {
									p, err := mailmsg.NextPart()
									if err == io.EOF || (err != nil && strings.Contains(strings.ToUpper(err.Error()), "EOF")) {
										break BodyPartCycle // 遍历结束
									} else if err != nil {
										continue BodyPartCycle
									}

									var _ct = strings.ToLower(p.Header.Get("Content-Type"))
									var _ed = strings.ToLower(p.Header.Get("Content-Transfer-Encoding"))
									_mt, _mtp, err := mime.ParseMediaType(_ct)
									if err != nil {
										continue BodyPartCycle
									}

									if charset, ok := _mtp["charset"]; ok && strings.ToLower(charset) != "utf-8" {
										continue BodyPartCycle
									}

									var data []byte

									switch _mt {
									case "text/html":
										fallthrough
									case "text/plain":
										data, err = io.ReadAll(p.Body)
										if err != nil {
											continue BodyPartCycle
										}
									default:
										continue BodyPartCycle
									}

									if _mt == "text/plain" || (mimeType != "text/plain" && _mt == "text/html") {
										contentType = _ct
										mimeType = _mt
										mimeParams = _mtp
										encoding = _ed
										bodyStr = string(data)
									}

									if mimeType == "text/plain" {
										break BodyPartCycle
									}
								}

								_ = mimeParams // 目前暂时不使用 mimeParams 这里写个语句防止 not use 保存
								_ = encoding   // 目前暂时不使用 encoding 这里写个语句防止 not use 保存

								if contentType == "" || mimeType == "" || bodyStr == "" {
									_ = errFunc("邮件无法被读取，我们只接受 text/plain 和 text/html")
									return // return msg read cycle
								}

								switch mimeType {
								case "text/plain":
									// ok
								case "text/html":
									data, err := html2text.FromString(bodyStr)
									if err != nil {
										_ = errFunc("邮件无法被读取，text/html 格式可能存在问题，无法被转换为纯文本，建议发送 text/plain 格式的邮件")
										return // return msg read cycle
									}
									bodyStr = data
								default:
									_ = errFunc("邮件无法被读取，我们只接受 text/plain 和 text/html")
									return // return msg read cycle
								}

								bodyStr = strings.ReplaceAll(bodyStr, "\r\n", "\n")
								bodyStr = strings.TrimLeft(bodyStr, "\n")
								bodyStr = strings.TrimRight(bodyStr, "\n")
								bodyStr = strings.TrimSpace(bodyStr)

								if bodyStr == "" {
									_ = errFunc("邮件内容为空")
									return // return msg read cycle
								}

								bodyStr, bodySafe = utils.ChangeDisplaySafeUTF8(bodyStr)
								if bodyStr == "" {
									_ = errFunc("邮件存在不安全因素")
									return // return msg read cycle
								}

								mailID := utils.GetIMAPMailID(messageID, senderAddr.String(), fromAddr.String(), toAddr.String(), replyAddr.String(), subject, bodyStr, messageDate, now)

								vxchan := make(chan bool, 2)
								emailchan := make(chan bool, 2)
								thankchan := make(chan bool, 2)

								go func(mailID string, messageID string, senderAddr string, fromAddr string, toAddr string, replyToAddr string, subject string, content string, date time.Time, t time.Time, vxchan chan bool, emailchan chan bool, thankchan chan bool) {
									defer func() {
										defer func() {
											_ = recover() // 防止chan关闭
										}()

										// 兜底
										vxchan <- false
										emailchan <- false
										thankchan <- false
									}()

									defer func() {
										if r := recover(); r != nil {
											if _err, ok := r.(error); ok {
												fmt.Printf("数据库提交消息出现致命错误: %s\n", _err.Error())
											} else {
												fmt.Printf("数据库提交消息出现致命错误（非error）: %v\n", r)
											}
										}
									}()

									err := sender.IMAPDataBase(mailID, messageID, senderAddr, fromAddr, toAddr, replyToAddr, subject, content, date, t)
									if err != nil {
										fmt.Printf("数据库提交消息出现错误: %s", err.Error())
										vxchan <- false
										emailchan <- false
										thankchan <- false
										return
									}

									vxchan <- true
									emailchan <- true
									thankchan <- true
								}(mailID, messageID, senderAddr.String(), fromAddr.String(), toAddr.String(), replyAddr.String(), subject, bodyStr, messageDate, now, vxchan, emailchan, thankchan)

								go func(vxchan chan bool) {
									defer close(vxchan)

									messageStr, err := func() (messageStr string, err error) {
										defer func() {
											if r := recover(); r != nil {
												if _err, ok := r.(error); ok {
													fmt.Printf("企业微信发送消息出现致命错误: %s\n", _err.Error())
													if err != nil {
														err = _err
														return
													}
												} else {
													fmt.Printf("企业微信发送消息出现致命错误（非error）: %v\n", r)
													if err != nil {
														err = fmt.Errorf("%v", r)
														return
													}
												}
											}
										}()

										var msgBuilder strings.Builder
										// 标准头部
										messageutils.WriteMessageStdHeader(&msgBuilder, "IMAP - 邮箱留言", mailID, now)

										msgBuilder.WriteString(fmt.Sprintf("主题: %s\n", subject))
										msgBuilder.WriteString(fmt.Sprintf("邮件 MessageID: %s\n", messageID))
										msgBuilder.WriteString(fmt.Sprintf("发送人: %s\n", utils.FormatEmailAddressToHumanStringMustSafe(senderAddr)))
										msgBuilder.WriteString(fmt.Sprintf("宣称发送人: %s\n", utils.FormatEmailAddressToHumanStringMustSafe(fromAddr)))
										msgBuilder.WriteString(fmt.Sprintf("回复地址: %s\n", utils.FormatEmailAddressToHumanStringMustSafe(replyAddr)))
										msgBuilder.WriteString(fmt.Sprintf("收件人: %s\n", utils.FormatEmailAddressToHumanStringMustSafe(toAddr)))
										msgBuilder.WriteString(fmt.Sprintf("邮件日期: %s %s\n", messageDate.Format("2006-01-02 15:04:05"), messageDate.Location().String()))

										if bodySafe {
											msgBuilder.WriteString(fmt.Sprintf("邮件内容是否安全：是\n"))
										} else {
											msgBuilder.WriteString(fmt.Sprintf("邮件内容是否安全：否，已处理\n"))
										}

										msgBuilder.WriteString(fmt.Sprintf("消息长度：%d\n", len(bodyStr)))

										const start = "---消息开始---\n"
										const stop = "\n---消息结束---"
										const splitTip = "以下消息的正文部分将开始分段发送\n"
										const splitAllTip = "以下新消息将开始分段发送\n"

										tmpMsg := msgBuilder.String()

										if len(tmpMsg)+len(start)+len(bodyStr)+len(stop) <= 2040 {
											msgBuilder.WriteString(fmt.Sprintf("%s%s%s", start, bodyStr, stop))

											messageStr = msgBuilder.String()

											err = sender.IMAPWechatRobot(messageStr)
											if err != nil {
												fmt.Printf("企业微信发送消息出现错误: %s\n", err.Error())
												return "", err
											}

											return messageStr, nil
										} else if len(tmpMsg)+len(splitTip) <= 2040 {
											msgBuilder.WriteString(splitTip)
											messageStr = msgBuilder.String()

											err = sender.IMAPWechatRobot(messageStr)
											if err != nil {
												fmt.Printf("企业微信发送消息出现错误: %s\n", err.Error())
												return "", err
											}

											sendMsg, err := sender.IMAPWechatRobotSplitMsg(bodyStr, mailID)
											if err != nil {
												fmt.Printf("企业微信发送消息出现错误: %s\n", err.Error())
												return "", err
											}

											msgBuilder.WriteString(sendMsg)
											messageStr = msgBuilder.String()

											return messageStr, nil
										} else {
											msgBuilder.WriteString(fmt.Sprintf("%s%s%s", start, bodyStr, stop))
											messageStr = splitAllTip + msgBuilder.String()

											sendMsg, err := sender.IMAPWechatRobotSplitMsg(messageStr, mailID)
											if err != nil {
												fmt.Printf("企业微信发送消息出现错误: %s\n", err.Error())
												return "", err
											}

											return sendMsg, nil
										}
									}()

									func(vxErr error) {
										if <-vxchan {
											_ = database.UpdateIMAPWxRobotSendMsg(mailID, messageStr, vxErr)
										}
									}(err)
								}(vxchan)

								go func(subject string, from string, now time.Time, emailchan chan bool) {
									defer close(emailchan)

									messageStr, smtpID, err := func() (messageStr string, smtpID string, err error) {
										defer func() {
											if r := recover(); r != nil {
												if _err, ok := r.(error); ok {
													fmt.Printf("邮件发送消息出现致命错误: %s\n", _err.Error())
													if err != nil {
														err = _err
														return
													}
												} else {
													fmt.Printf("邮件发送消息出现致命错误（非error）: %v\n", r)
													if err != nil {
														err = fmt.Errorf("%v", r)
														return
													}
												}
											}
										}()

										var msgBuilder strings.Builder
										// 标准头部
										messageutils.WriteMessageStdHeader(&msgBuilder, "IMAP - 邮箱留言", mailID, now)

										msgBuilder.WriteString(fmt.Sprintf("主题: %s\n", subject))
										msgBuilder.WriteString(fmt.Sprintf("邮件 MessageID: %s\n", messageID))
										msgBuilder.WriteString(fmt.Sprintf("发送人: %s\n", utils.FormatEmailAddressToHumanStringMustSafe(senderAddr)))
										msgBuilder.WriteString(fmt.Sprintf("宣称发送人: %s\n", utils.FormatEmailAddressToHumanStringMustSafe(fromAddr)))
										msgBuilder.WriteString(fmt.Sprintf("回复地址: %s\n", utils.FormatEmailAddressToHumanStringMustSafe(replyAddr)))
										msgBuilder.WriteString(fmt.Sprintf("收件人: %s\n", utils.FormatEmailAddressToHumanStringMustSafe(toAddr)))
										msgBuilder.WriteString(fmt.Sprintf("邮件日期: %s %s\n", messageDate.Format("2006-01-02 15:04:05"), messageDate.Location().String()))

										if bodySafe {
											msgBuilder.WriteString(fmt.Sprintf("邮件内容是否存在不安全因素：不存在不安全因素\n"))
										} else {
											msgBuilder.WriteString(fmt.Sprintf("邮件内容是否存在不安全因素：存在不安全因素，已被移除\n"))
										}

										msgBuilder.WriteString(fmt.Sprintf("消息长度：%d\n", len(bodyStr)))
										msgBuilder.WriteString(fmt.Sprintf("---消息开始---\n%s\n---消息结束---", bodyStr))

										messageStr = msgBuilder.String()

										smtpID, err = sender.IMAPEmail(subject, from, messageStr, now)
										if err != nil {
											fmt.Printf("邮件发送消息出现错误: %s\n", err.Error())
											return "", smtpID, err
										}

										return messageStr, smtpID, nil
									}()

									func() {
										if <-emailchan {
											_ = database.UpdateIMAPEmailSendMsg(mailID, messageStr, smtpID, err)
										}
									}()
								}(subject, utils.FormatEmailAddressToHumanStringJustNameMustSafe(fromAddr), now, emailchan)

								go func(subject string, messageID string, fromAddr *mail.Address, toAddr *mail.Address, replyAddr *mail.Address, thankchan chan bool) {
									defer close(thankchan)

									smtpID, err := smtpserver.SendThankMsg(subject, messageID, fromAddr, toAddr, replyAddr)

									if <-thankchan {
										_ = database.UpdateIMAPThankEmailSendMsg(mailID, smtpID, err)
									}
								}(subject, messageID, fromAddr, toAddr, replyAddr, thankchan)
							}()

							processSeqSet.AddNum(msg.SeqNum)
							processSeqLen += 1
						}

						err = msgCMD.Close()
						if err != nil {
							fmt.Printf("close fetch failed: %s\n", err.Error())
							return // return msg search cycle
						}

						if processSeqLen > 0 {
							err = imapClient.Store(processSeqSet, &imap.StoreFlags{
								Op:    imap.StoreFlagsAdd,
								Flags: []imap.Flag{imap.FlagSeen},
							}, nil).Close()
							if err != nil {
								fmt.Printf("close store failed: %s\n", err.Error())
								return // return msg search cycle
							}
						}
					}()

					isPass := func() bool {
						var idle *imapclient.IdleCommand
						var sleepTime = DefaultImapCycleTime

						defer func() {
							if supportIDLE && idle == nil || !supportIDLE && idle != nil {
								fmt.Printf("客户端 对IDLE操作出错 停用IDLE")
								supportIDLE = false
							}

							if idle != nil {
								_ = idle.Close()
							}
						}()

						if supportIDLE {
							_idle, err := imapClient.Idle()
							if err != nil {
								idle = nil
								sleepTime = DefaultImapCycleTime
								supportIDLE = false
							} else {
								idle = _idle
								sleepTime = DefaultImapCycleWithIdleTime
								supportIDLE = true
							}
						} else {
							idle = nil
							sleepTime = DefaultImapCycleTime
							supportIDLE = false
						}

						var noopstopchan chan bool = nil
						var noopstopokchan chan bool = nil

						if !supportIDLE || idle == nil {
							// 覆写一遍，纠错
							idle = nil
							sleepTime = DefaultImapCycleTime
							supportIDLE = false

							noopstopchan = make(chan bool, 1)
							noopstopokchan = make(chan bool, 1)

							go func(noopstopchan chan bool, noopstopokchan chan bool) {
								defer func() {
									noopstopokchan <- true
								}()

								noopTicker := time.NewTicker(DefaultImapNoopTime)
								defer noopTicker.Stop()

							NoopCycle:
								for {
									select {
									case <-noopTicker.C:
										_ = imapClient.Noop().Wait()
									case <-noopstopchan:
										break NoopCycle
									}
								}
							}(noopstopchan, noopstopokchan)
						}

						res := func() bool {
							defer func() {
								if supportIDLE && idle != nil && noopstopchan != nil && noopstopokchan != nil {
									noopstopchan <- true
									<-noopstopokchan
								}
							}()

							select {
							case <-unilateralData:
								return true
							case <-time.After(sleepTime):
								return true
							case <-stopchan:
								return false
							}
						}()

						return res
					}()

					if !isPass {
						return StatusStop
					} else if imapClient.Noop().Wait() != nil {
						// 连接中断
						return StatusConnectBlock
					}
				}
			}()

			if status == StatusConnectError {
				time.Sleep(10 * time.Second)
			} else if status == StatusConnectBlock {
				// 马上循环重新连接
			} else if status == StatusCommandError {
				commandErrorCount += 1
				if commandErrorCount > 10 {
					notifySubject := fmt.Sprintf("【%s】系统通知", resource.Name)
					notifyContent := fmt.Sprintf("连接 IMAP 服务器出现多次命令错误，超过10次重启 IMAP 服务，现将要停止 IMAP 服务。")
					systemnotify.SendNotify(notifySubject, notifyContent)
					break MainCycle
				} else {
					time.Sleep(15 * time.Second)
				}
			} else if status == StatusStop {
				break MainCycle
			}
		}

		fmt.Printf("IMAP 服务结束\n")
	}()

	return stopchan, nil
}

func isTemporaryNetError(err error) bool {
	// 检查是否为超时错误
	if errors.Is(err, net.ErrClosed) {
		return true
	}
	// 可以添加更多你认为应该重试的错误类型检查
	// 例如，如果 err 是一个 *net.OpError，并且它的 Temporary() 返回 true，那也表示是一个暂时性错误。
	// 注意根据实际情况调整这里的错误处理逻辑。
	return false
}
