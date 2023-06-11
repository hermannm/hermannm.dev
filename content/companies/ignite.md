---
name: Ignite Procurement
slug: ignite
iconPath: /img/companies/ignite.webp
iconAlt: "Ignite Procurement company logo"
techStack:
  - tech: TypeScript
    usedFor: frontend
    usedWith:
      - React
  - tech: Go
    usedFor: backend microservices
  - tech: gRPC
  - tech: GraphQL
linkCategories:
  - title: To read more about Ignite
    links:
      - text: igniteprocurement.com
        link: https://www.igniteprocurement.com/
---

The story of this project starts in the fall of 2020, when my dad gave me one of the coolest
birthday gifts I've ever received. He made me my very own board game: _The Battle for Hermannia_!
The game involves strategy, diplomacy and intrigue, and I've played it a lot with friends in
Trondheim. A couple friends of mine were particular fans of the game, and so in November 2021, we
started developing a digital edition of it as a hobby project.

We decided to build the digital edition as a proper multiplayer online game. I wrote the server in
Go, finding its native concurrency support suitable for the parallel nature of the game. My friends
and I collaborated on the client, where we used C# with the Unity game engine. During this time, one
of my friends also founded [Immerse NTNU](https://immersentnu.no/), a student organization for game
development. He made this game client their first project, and so we also got invaluable help from
new members there.
