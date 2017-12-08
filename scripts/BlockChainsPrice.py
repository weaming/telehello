#!/usr/bin/env python
# coding: utf-8
"""
Author       : weaming
Created Time : 2017-08-17 01:44:46

File Name    : BlockChainsPrice.py
Description  :
    https://www.jubi.com/help/api.html
"""
import json
import requests
import yaml
import sys

API_BASE = 'https://blockchain.info/ticker'
EXCHANGE_API = 'http://api.fixer.io/latest?symbols=CNY&base=USD'
INBOX = 'http://aws.bitsflow.org:1234/api/new'
coins = ('btc', 'bcc', 'ltc', 'eth', 'etc')

results = []

# BTC price
#for c in coins:
res = requests.get(API_BASE, headers={'User-Agent': 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.79 Safari/537.36'})
if res.status_code == 200:
    try:
        data = json.loads(res.text)['CNY']
    except:
        print res.text
        sys.exit(1)
    # only leave 'last' and 'vol'
    with open('/var/btc', 'r') as f:
        yesterday = float(f.read())
    with open('/var/btc', 'w') as f:
        f.write(str(data['last']))
    data['yesterday'] = yesterday          
    data['now'] = data['last']
    for k in ('15m', 'buy', 'sell', 'symbol', 'last'):
        if k in data:
            del(data[k])
    output = yaml.safe_dump(data, default_flow_style=False)
    results.append('%s:\n%s' % ('BTC'.upper(), output))
else:
    print res.status_code

# USD exchange
res = requests.get(EXCHANGE_API, headers={'User-Agent': 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.79 Safari/537.36'})
if res.status_code == 200:
    try:
        data = json.loads(res.text)['rates']
    except:
        print res.text
        sys.exit(1)
    output = yaml.safe_dump(data, default_flow_style=False)
    results.append('%s:\n%s' % ('USD', output))
else:
    print res.status_code

# sending
if len(results) > 0:
    requests.post(INBOX, '\n'.join(results))
