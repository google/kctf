#!/usr/bin/env python3

from winrm.protocol import Protocol
import sys
import re
import os

with open(os.environ.get('CREDS_PATH'), 'r') as fd:
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

print(run_cmd(sys.argv[1:]).decode('ascii'))
