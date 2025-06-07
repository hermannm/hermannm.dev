---
name: casus-belli
path: /casus-belli
tagLine: Online multiplayer board game.
goPackage:
  rootName: hermannm.dev/casus-belli
  githubURL: https://github.com/hermannm/casus-belli
logo:
  path: /img/logos/casus-belli.png
  altText: A fort, surrounded by forest
techStack:
  - tech: Go
    usedFor: server-side
  - tech: Godot
    usedFor: client-side
    usedWith:
      - C#
links:
  - title: Code
    link: https://github.com/hermannm/casus-belli
---

The story of this project started in the fall of 2020, when my dad gave me the coolest birthday gift
I've ever received: he made me my very own board game! The game involves strategy, diplomacy and
battle, and I've played it a lot with friends and family. A couple friends of mine were particular
fans of the game, so in the fall of 2021, we started developing a digital edition of it as a hobby
project. Thus, _Casus Belli_ was born.

We decided to build the digital edition as a proper online multiplayer game. I wrote the server in
Go, finding its native concurrency support suitable for the parallel nature of the game. My friends
and I collaborated on the client, where we initially used the Unity game engine. During this time,
one of my friends also founded [Immerse NTNU](https://immersentnu.no/), a student organization for
game development. We continued the project under that organization, and so we also got great
contributions from new members there.

In the fall of 2023, the company behind Unity decided to
[upend the terms for developers using their game engine](https://blog.unity.com/news/plan-pricing-and-packaging-updates).
Although this change likely would not affect our project, I found Unity's practices here quite
abhorrent, and it gave me a distaste for using the engine further. Since we were developing this
game as an open-source project, it felt more appropriate to also use open-source tools for it. Thus,
we decided to make the switch over to Godot, an open-source game engine.
