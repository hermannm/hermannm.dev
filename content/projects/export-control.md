---
name: export-control
slug: export-control
tagLine: Search tool for the Norwegian Export Control Law.
logoPath: /img/logos/safetec.png
logoAlt: Safetec company logo
techStack:
  - tech: Python
    usedWith:
      - Django
linkGroups:
  - title: Code
    links:
      - text: cdp-group4/export-control
        link: https://github.com/cdp-group4/export-control
---

In the fall semester of 2022, I took the course _Customer-Driven Project_ at NTNU. In the course,
five other Computer Science students and I were given a project from a real customer: _Safetec_, a
consulting firm specializing in risk management. Our project was to build a tool for searching
through the regulations of the Norwegian Export Control Law, which restricts the export of certain
sensitive products and technologies. Just a year before our project, the regulations were expanded,
and now included a 250 page list with a wide variety of regulated items. Several of Safetec's
customers were worried about how the regulations might affect them, and had trouble navigating the
new lists. The goal of our tool was to be a resource that Safetec could give to their customers, so
they can find out if export control affects them.

Our final product was a web application built with Django, using PostgreSQL for its database. We
deployed the application in Safetec's Azure cloud, where it is now running at
[exportcontrol.safetec.no](https://exportcontrol.safetec.no/). I was quite proud of our final
result, as we were able to deliver a fully functional product that the customer was happy with. The
project taught me a lot about the back-and-forth process of requirement elicitation, and dealing
with changing requirements.

![The Export Control search interface](/img/screenshots/export-control.png)
