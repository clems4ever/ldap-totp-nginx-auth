---
layout: default
title: Push Notification
parent: Second Factor
nav_order: 3
grand_parent: Features
---

# Mobile Push Notification

Mobile push notifications is the new trendy second factor method. When second factor is requested
by Authelia, a notification is sent on your phone that you can either accept or deny.

<p align="center">
  <img src="../../images/duo-push-1.jpg" width="200">
  <img src="../../images/duo-push-2.png" width="200">
</p>


Authelia leverages [Duo] third party to provide this feature.

First, sign up on their website, log in, create a user account and attach it a mobile device.
Beware that the name of the user must match the name of the user in Authelia.

Then, in Duo interface, click on *Applications* and *Protect an Application*. Select the option
*Partner Auth API*. This will generate an integration key, a secret key and a hostname. You can
set the name of the application to **Authelia** and then you must add the generated information
to Authelia [configuration](../../deployment/index.md) as shown below:

    duo_api:
      hostname: api-123456789.example.com
      integration_key: ABCDEF
      secret_key: 1234567890abcdefghifjkl

Now that Authelia is configured, pass the first factor and select the Push notification
option.

<p align="center">
  <img src="../../images/2FA-PUSH.png" width="400">
</p>

You should now receive a notification on your mobile phone with all the details
about the authentication request.
In case you have multiple devices available, you will be asked to select your preferred device.


## Limitation

Users must be enrolled via the Duo Admin panel, they cannot enroll a device from
**Authelia** yet.


## FAQ

### Why don't I have access to the *Push Notification* option?

It's likely that you have not configured **Authelia** correctly. Please read this
documentation again and be sure you had a look at [config.template.yml](https://github.com/authelia/authelia/blob/master/config.template.yml).

[Duo]: https://duo.com/
