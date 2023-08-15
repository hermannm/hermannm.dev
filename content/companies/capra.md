---
name: Capra
slug: capra
iconPath: /img/companies/capra.webp
iconAlt: "Capra company logo"
techStackTitle: Technologies I worked with
techStack:
  - tech: TypeScript
    usedFor: frontend
    usedWith:
      - React
  - tech: Kotlin
    usedFor: backend
  - tech: AWS CDK
    usedFor: infrastructure-as-code
linkGroups:
  - title: More about Capra
    links:
      - text: capraconsulting.no
        link: https://www.capraconsulting.no/
---

In June 2023, I started a summer internship as a software developer at Capra, an IT consulting firm.
I worked in the 'Liflig' department, which takes more of an in-house approach to their projects,
with greater control of the tech stack and deployment. Our project this summer was to build a
backend system and admin platform for [IDTAG](https://www.idtagtech.com/), an exciting startup!

Liflig intended to continue working with IDTAG as a customer after our summer project, which meant
that we had to emphasize test coverage, code quality and robustness in our implementation. While
this did add some pressure on us, I found it much more rewarding to work on something where our
decisions had real impact, rather than a mere proof-of-concept that I've heard are common in other
internships. In the end, we developed a solid test suite with both integration and unit tests. I was
especially happy with our integration test setup, which used
[testcontainers](https://testcontainers.com/) and [LocalStack](https://localstack.cloud/) to run our
tests against a real database and AWS API.

On the frontend, I got the opportunity to work with [Tailwind CSS](https://tailwindcss.com/), and
found it excellent for developer productivity. Styling with Tailwind resembled how I normally do
things with vanilla CSS, namely composing a bunch of utility classes together to build up the UI.
With Tailwind, I no longer had to spend time naming classes, and avoided the pitfall of forgetting
to delete unused classes from a CSS file.

In the last week of the project, I implemented a backend API for
[Apple OAuth login](https://developer.apple.com/documentation/sign_in_with_apple/sign_in_with_apple_rest_api),
used by IDTAG's existing iOS app. This was quite the challenge, and taught me a lot about the
difficulties of interacting with external APIs. Though it was frustrating to figure out poor error
messages from Apple's API, the satisfaction of making it work in the end made it all worth it.
