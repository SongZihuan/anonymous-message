# 使用Git Bash运行
# 需要先安装<chocolatey>
# 安装命令：iex ((New-Object System.Net.WebClient).DownloadString('https://chocolatey.org/install.ps1'))
# 若安装过程提示无权限执行脚本，则：Set-ExecutionPolicy Bypass -Scope Process -Force
# 并且删除<C:\ProgramData\chocolatey>目录。
# 安装完成后使用<choco --version>查看版本，确定安装成功
# 安装<dos2unix>：choco install dos2unix

find . -type f -name '*' -print0 | xargs -0 dos2unix