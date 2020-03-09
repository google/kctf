# Copyright 2020 Google LLC
# 
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
# 
#     https://www.apache.org/licenses/LICENSE-2.0
# 
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
FROM kctf-healthcheck

# Add the exploit directory as an encrypted archive.
# This is out of caution in case that we leak the 
COPY .gen/exploit.cpio.enc /

COPY files /home/user

CMD decrypt_exploit.sh && setpriv --init-groups --reuid user --regid user --inh-caps=-all -- /exploit/run.sh & /home/user/healthz.py
