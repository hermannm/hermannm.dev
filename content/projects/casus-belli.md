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

The story of this project started in the fall of 2020, when my dad gave me one of the coolest gifts
I've ever received: he made me my very own board game! The game involves strategy, diplomacy and
battle, and I've played it a lot with friends and family. A couple friends of mine were particular
fans of the game, so in the fall of 2021, we started developing a digital edition of it as a hobby
project. Thus, _Casus Belli_ was born.

We decided to build the digital edition as a full-fledged online multiplayer game. I wrote the
server in Go, finding its native concurrency support suitable for the parallel nature of the game.
My friends and I collaborated on the client, where we initially used the Unity game engine. But in
the fall of 2023, the company behind Unity decided to
[upend the terms for developers using their game engine](https://blog.unity.com/news/plan-pricing-and-packaging-updates).
I found Unity's practices here quite distasteful, and so we decided to switch the client over to the
open-source Godot game engine.

Once I started working full-time in 2024, progress on the game stalled, as it was difficult to find
energy for it in my spare time while I was programming all day at work. I still want to finish the
game at some point though. The implementation of the server is basically finished, and work on the
client is well on its way. Hopefully, I'll be able to finish the game some time in the next couple
of years, and release it to the public.

![The digital edition of the Casus Belli board](/img/screenshots/casus-belli.png)
