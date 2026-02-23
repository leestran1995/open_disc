# open_discord

Discord without the privacy concerns, someday. It's really just an IRC client right now. If we can get voip
working someday, then we'd really be cooking.

## Overview

open_discord is intended to be self-hosted rather than centralized. Each server is a completely separate
instance, so a user's username/password on `Server A` is completely separate from `Server B`.

At the moment, all users in a server have access to all rooms in that server.

## Feature plans

- Immediate next thing:
  - Rescope messages -> ServerEvents
  - Messages, users connecting/disconnecting, rooms being created, etc. are all ServerEvents
  - ServerEvents are all stored on a single table with an auto-incrementing column
  - Clients will use that auto-inc column for two purposes
    - To know if they missed any events
    - To know where the last event they received was, to re-sync with the server
- ~~Room Ordering~~
- Username color customization
- Room Categories
- Configurable nicknames
- Server commands
  - For example, to configure nicknames
- User roles
  - Only users with certain roles can see certain channels
  - Only users with certain roles can create new channels
  - If a user does not have the role, they will receive ServerEvents with a type of [REDACTED] to maintain the auto-inc sequence
- Threads
  - Slack-style
- VOIP
- Images uploads
- Data retention policies
  - Automatically clean up data older than some configured timeframe
- Rate Limiting
- User Presence
  - Use Redis to display whether a user is actively connected or not

## Implementation Plans

- Refactor rooms to each be run in its own goroutine
- Rework `message` table to be scoped to any and all `RoomEvents`
- 

## Server Event Architecture & Approach

### Event Order
ServerEvents will contain an incrementing event_order integer. This will allow the FE to know if it missed any events
(for example, if the FE receives event 1000 followed by 1002, it knows it missed 1001) and request the missing events.
It also provides an easy way for the FE to say "The last event I saw was _, give me everything since then"

### Event Payloads
Identifiers in server event payloads should be the unchanging unique identifier (uuid) for the related domain object.
For optimization's sake, we might consider sending changeable data (for example, user nicknames in messages) along with the
events to save the FE the need to do a lookup.

## Jargon

- UserID
  - UUID unique identifier for a given user, safe to send in API Responses
- Username
  - The username a user uses to log in to the server. Best not to include in API Responses for security's sake
- Nickname
  - User-chosen nickname that will be displayed in the frontend
- Server Event
  - Event representing something that has happened in the server, such as a message being sent or a new user joining the server.
- Room
  - Text channel

## Auth

The whole point of this project is to avoid interacting with giant companies that don't care about user privacy
(and coincidentally provide oauth services) so we're rolling our own username/password based verification.

2FA would be cool to implement, at some point.

### Signup

![img.png](documentation/images/signup_flow.png)

### Login

After the user is logged in, they will send their JWT to all subsequent API requests.
We will use the username embedded in the JWT to determine which user sent the API request.

To prevent users from needing to re-login once the JWT expires, eventually we'll want to implement
refresh tokens.

![img.png](documentation/images/login_flow.png)

## A Note on AI Usage

The backend was coded entirely by hand by me, https://github.com/leestran1995. That being said, GoLand's
autocomplete is really good, so a lot of the `if err != nil { return err }` checks were just tabbed through, but I'll
be dead in a ditch before I cheat on a portfolio project.

The frontend was coded
by https://github.com/vrennat by way of Claude.
