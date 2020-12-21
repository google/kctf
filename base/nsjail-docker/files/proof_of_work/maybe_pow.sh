#!/bin/bash
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

POW_FILE="/kctf/pow/pow.conf"

if [ -f ${POW_FILE} ]; then
  source /venv/bin/activate

  POW="$(cat ${POW_FILE})"
  if ! /usr/bin/pow.py ask "${POW}"; then
    echo 'pow fail'
    exit 1
  fi
fi

exec "$@"
