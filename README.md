Image Generator
===

Generates images for bbbgp.org; lives at https://images.bbbgp.org.

It uses random images from files uploaded to the site and
calculates out the next meeting time to generate a poster.


Routes
---
/upload
/meeting
/cancelled
/pizza

Different URL pathways have different images generated for them.
Default is: https://images.bbbgp.org/meeting

The password for upload is set as an environmental variable "password".
