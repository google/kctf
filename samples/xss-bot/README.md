Example XSS Bot

This is based on the challenge skeleton with a few modifications:
* Disable nsjail in the cmd line (Dockerfile) and remove the nsjail config
* Create a tmpfs at `/tmp` via `config/advanced/deployment/containers.yaml`
* Modify the Dockerfile to install puppeteer
* Replace the flag with a cookie
* Add the puppeteer script in `challenge/image/bot.js`
* Implement a healthcheck
