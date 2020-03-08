FROM kctf-healthcheck

# Add the exploit directory as an encrypted archive.
# This is out of caution in case that we leak the 
COPY .gen/exploit.cpio.enc /
COPY files /home/user

CMD decrypt_exploit.sh && setpriv --init-groups --reuid user --regid user --inh-caps=-all -- /exploit/run.sh & /home/user/healthz.py
