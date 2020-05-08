#!/usr/bin/env python2.7

from __future__ import print_function
from winrm.protocol import Protocol
import sys
import subprocess
import re

GS_PATH, IMAGE_TAG = sys.argv[1:]

with open('windows_creds', 'r') as fd:
    hostname = re.match('ip_address:\s*(.*)', fd.readline()).group(1)
    password = re.match('password:\s*(.*)', fd.readline()).group(1)
    username = re.match('username:\s*(.*)', fd.readline()).group(1)

p = Protocol(
    endpoint='https://{}:5986/wsman'.format(hostname),
    username=username,
    password=password,
    server_cert_validation='ignore')
shell_id = p.open_shell()

def run_cmd(cmd, check=True):
    print("running cmd {}".format(cmd), file=sys.stderr)
    command_id = p.run_command(shell_id, cmd[0], cmd[1:])
    std_out, std_err, status_code = p.get_command_output(shell_id, command_id)
    p.cleanup_command(shell_id, command_id)
    if check and status_code != 0:
        print('status: {}'.format(status_code), file=sys.stderr)
        print('=== stdout ===', file=sys.stderr)
        print(std_out, file=sys.stderr)
        print('=== stderr ===', file=sys.stderr)
        print(std_err, file=sys.stderr)
        exit(1)
    return std_out

run_cmd(['rmdir', 'c:\\build', '/s', '/q'], check=False)
run_cmd(['mkdir', 'c:\\build'])
run_cmd(['gsutil', 'cp', '-r', GS_PATH+'/*', 'c:\\build'])
run_cmd(['gcloud', 'auth', 'configure-docker', '--quiet'])
tmp_tag = '{}:tmp'.format(IMAGE_TAG)
run_cmd(['docker', 'build', '-t', tmp_tag, 'c:\\build'])
digest = run_cmd(['docker', 'image', 'ls', '-q', tmp_tag]).strip()
full_tag = '{}:{}'.format(IMAGE_TAG, digest)
run_cmd(['docker', 'tag', tmp_tag, full_tag])
run_cmd(['docker', 'push', full_tag])
p.close_shell(shell_id)
print(full_tag, end='')
