---
name: coffeetalk
slug: coffeetalk
tagLine: Peer-to-peer video chat application.
logoPath: /img/logos/coffeetalk.png
logoAlt: CoffeeTalk project logo
techStack:
  - tech: JavaScript
    usedFor: frontend
  - tech: Go
    usedFor: backend
  - tech: WebRTC
  - tech: MQTT
linkGroups:
  - title: Code
    links:
      - text: dcs-team4/coffeetalk
        link: https://github.com/dcs-team4/coffeetalk
---

In the spring semester of 2022, I took the course _Design of Communicating Systems_ at NTNU. A part
of that course was a group project to build a video chat application. And thus CoffeeTalk was born:
a web app with peer-to-peer video streaming, and even a basic quiz game! We developed the
application as 3 servers (one for serving the web app, one for WebRTC video stream coordination, and
one for the quiz game using MQTT). The servers were all written in Go, while the web app was written
in vanilla JavaScript.

In the end, we deployed the app with [DigitalOcean](https://www.digitalocean.com/). My proudest
moment that semester was having an hour-long conversation with my dad through my own video chat app!

![My dad and I talking through CoffeeTalk](/img/screenshots/coffeetalk.png)
