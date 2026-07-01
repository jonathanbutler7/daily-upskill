# Helpful links

- https://gist.github.com/vasanthk/485d1c25737e8e72759f
- https://bytebytego.com/guides/system-design-cheat-sheet/
- https://github.com/donnemartin/system-design-primer
- https://github.com/ashishps1/awesome-system-design-resources?tab=readme-ov-file
- https://paperdraw.dev/
- https://www.youtube.com/shorts/Ibf7wKf8MqU

System design interviews often start with the choice about CAP theorem.
But the real decision is a lot simpler than you would think.
First, a really quick refresher. CAP theorem stands for consistency, availability, and partition tolerance.
The classic framing is that a distributed system can only guarantee two of the three. But in a distributed system, the network will eventually fail. That's a guarantee. And so partition tolerance is not optional.
This means that CAP theorem isn't asking you to pick two of three. It's actually just asking one simple question.
When the network does fail, do you return possibly stale data, which would mean choosing availability, or do you return an error, which would mean choosing consistency?
Let's look at an example.
Say your user updates their profile in the United States, and the connection to your Europe server goes down before that update is replicated. Then, a user in Europe requests the profile.
Well, you have two options. You can show them the old picture, or you can refuse to answer until the network heals.
For a social app, you can clearly show them the old picture. Stale data is fine. An error here is not. That's choosing availability.
For Ticketmaster, on the other hand, selling the last seat at a concert, you would return an error.