# Rascal 

## What is Rascal

Rascal is a bot to facilitate IRC federation (using the IRCv3 `draft/relaymsg` spec).

## Misc Notes

### General Usage

User in the channel issues `rascal: user/some.irc.server.example` in the current channel (or `/msg rascal invite user/SISE #channel`)

If some.irc.server.example has never before been seen, then the bot returns an error that the server must be registered first.

> The server is federated with `/msg rascal federate SISE some.irc.server.example`

User in channel then issues `rascal: federate some.irc.server.example` or `/msg rascal federate some.irc.server.example #channel`

On the other server, Rascal creates a channel (if the server is known and has the `autocreate` option set) `#fed-<channel_name>` and invites the user to the channel. Rascal will then relay messages back and forth between the channels.

For manual configuration, the invited server can list the federation advertisements from other servers with:

```
/msg rascal advertisements <some.other.irc.server>

#a
#list
#of
#federated
#channel
```

An existing channel (where @rascal is present and opped) can then be configured to federate with `/msg rascal connect SOIS/#remote-channel #local-channel`.

The list of federated servers can be listed with:

`/msg rascal servers`

which produces a list like:

`SISE some.irc.server.example`
`SOIS some.other.irc.server`
`...`

Federated channels can be listed with:

`/msg rascal channels`

`#local-channel SISE,SOIS`
`...`

### Discovery

Server operators running Rascal can publish their IRC server to the (centralized) federation discovery system with:

`/msg rascal publish-to-discovery rascal/my.irc.server.example:6697 "This is a description of my server, who we are, what we like to talk about"`

Rascal bot will submit an api call to `POST https://rascal.noauthority.social/api/v1/discovery` and include the given information.

The discovery server will query the `rascal` user at the `my.irc.server.example:6697` to identify server feature-set, ensure the `rascal` version in use is compatible, etc.

The server `my.irc.server.example:6697` will then be added to the list of IRC servers offered by `GET https://rascal.noauthority.social/api/v1/discovery`, available within IRC by issuing `/msg rascal known-network`

**NOTE**: only 1 server per root domain and 1 server as a subdomain is permitted. e.g., you can register `foo.bar.example` to the Discovery service, and you can issue `bar.example` to the discovery service. However, attempts to then register `fizz.bar.example` will fail until you issue `/msg rascal unpublish-to-discovery rascal/foo.bar.example:6697`. This prevents a DoS on the discovery server via programmatic subdomain registration.

**NOTE:** `/msg rascal known-network` will include all known federatable servers: discovery server list + existing federated servers.


### Peering

Rascal bots will peer to other bots using IRC for signaling and control, however the message delivery system is capable of out-of-band delivery as well (ex, via websockets). Messages from the channel are batched and sent to subscribing servers with a (pre-negotiated) `HMAC_SHA256`.



