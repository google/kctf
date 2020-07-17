FROM mcr.microsoft.com/windows/servercore:1909-amd64

RUN curl.exe -L -o "C:\\python-installer.exe" https://www.python.org/ftp/python/3.8.4/python-3.8.4-amd64.exe
RUN C:\\python-installer.exe /passive InstallAllUsers=1 DefaultAllUsersTargetDir="C:\\python" Include_doc=0 PrependPath=1 Include_test=0

RUN C:\\python\\python.exe -m pip install requests

RUN mkdir C:\\app
COPY healthcheck.py C:\\app

EXPOSE 45281

ENTRYPOINT ["C:\\python\\python.exe", "C:\\app\\healthcheck.py"]
