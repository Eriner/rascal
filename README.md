# Rascal 

## What is Rascal

Rascal is a bot to facilitate IRC federation (using the IRCv3 `draft/relaymsg` spec).

## Misc Notes

The federation flow is as follows:

User in the channel issues `rascal: user/some.irc.server.example` in the current channel

If some.irc.server.example has never before been seen, then the bot returns an
error that the server must be registered first.

The server is registered with `/msg rascal federate SISE some.irc.server.example`

User in channel then issues `rascal: user/some.irc.server.example` (OR `user/SISE`)

Messages from that user will appear as user/SISE in either case. (A short-name is required during server registration)

The list of federated servers can be listed with:

`/msg rascal federated`

which produces a list like:

`SISE some.irc.server.example`
`SOIS some.other.irc.server`
`...`


Server operators running Rascal can publish their IRC server to the (centralized) federation discovery system with:

`/msg rascal publish-to-discovery rascal/my.irc.server.example:6697 "This is a description of my server, who we are, what we like to talk about"`

Rascal bot will submit an api call to `POST https://rascal.noauthority.social/api/v1/discovery` and include the given information.

The discovery server will query the `rascal` user at the `my.irc.server.example:6697` to identify server feature-set, ensure the `rascal` version in use is compatible, etc.

The server `my.irc.server.example:6697` will then be added to the list of IRC servers offered by `GET https://rascal.noauthority.social/api/v1/discovery`, available within IRC by issuing `/msg rascal known-network`

**NOTE**: only 1 server per root domain and 1 server as a subdomain is permitted. e.g., you can register `foo.bar.example` to the Discovery service, and you can issue `bar.example` to the discovery service. However, attempts to then register `fizz.bar.example` will fail until you issue `/msg rascal unpublish-to-discovery rascal/foo.bar.example:6697`. This prevents a DoS on the discovery server via programmatic subdomain registration.


**NOTE:** `/msg rascal known-network` will include all known federatable servers: discovery server list + existing federated servers.
