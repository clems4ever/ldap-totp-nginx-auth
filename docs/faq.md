---
layout: default
title: FAQ
nav_order: 8
---

# Frequently asked questions

## How can the backend be aware of the authenticated users?

This question is solved
[here](https://www.authelia.com/docs/deployment/supported-proxies/#how-can-the-backend-be-aware-of-the-authenticated-users).

## Why only use a private issue key with OIDC?

The reason for using only the private key here is that one is able to calculate the public key easily from the private
key (`openssl rsa -in rsa.key -pubout > rsa.pem`).
