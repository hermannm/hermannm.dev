---
page:
  goPackage:
    rootName: hermannm.dev/analysis
    githubURL: https://github.com/hermannm/analysis
name: analysis
slug: analysis
tagLine: Data analysis service, built for my master's thesis.
logoPath: /img/projects/analysis.png
logoAlt: Combined logos of ClickHouse and Elasticsearch
techStack:
  - tech: Go
  - tech: ClickHouse
  - tech: Elasticsearch
linkGroups:
  - title: Code
    links:
      - text: hermannm/analysis
        link: https://github.com/hermannm/analysis
---

In 2018, I started studying
[Industrial Economics and Technology Management](https://www.ntnu.edu/studies/mtiot) at NTNU
Trondheim. At the time, I was interested in both economics and engineering, so I found the degree to
be a good balance between the two. But as time went on, my interest in economics decreased, while my
interest in programming grew and grew. This came to a head in my fourth year, when I decided to
switch my degree to Computer Science. This extended my studies by half a year, but in return I got
to take technical courses that I truly loved, and write the master's thesis that I wanted.

The year prior to my thesis, I worked as a software developer for [Ignite](/ignite). After my summer
internship there, Ignite's CTO suggested to me a possible thesis one could write about the Ignite
platform. You see, Ignite uses Elasticsearch to enable complex data analytics for their customers.
This works well, but they've had issues when it comes to the performance of ingesting data, stale
reads, correctness of results and more. It would be interesting to study the analytical database
landscape, to see if other databases with different tradeoffs might be better suited for the type of
platform that Ignite has built. Thus, my thesis was born: "Replacing Elasticsearch in a Data
Analytics Platform".

I started out by searching for possible alternatives to Elasticsearch, and found that the world of
databases was even more vast than I expected. But eventually, I did find one particularly promising
database: ClickHouse, a column-oriented database that promises efficient analytical queries. I then
built _analysis_, a backend service with an analytical query API, where you can toggle between using
Elasticsearch or ClickHouse as the backing database. With this, I could measure differences between
the two for different types of workloads. In the end, I found ClickHouse to be a viable alternative
to Elasticsearch, both in quantitative and qualitative aspects, though some mixed results and
limitations of the experiment made it not an obvious choice.
