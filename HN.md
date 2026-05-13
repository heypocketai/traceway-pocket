# A note from the creator of Traceway

I'm the creator of Traceway. I didn't make the HN post and only just realized it was up, so this file is my response to the thread since HN has been rate limiting my account and I can't reply to everyone there.

A lot of tools in this space, most pretty good. The goals when I started Traceway was:
- simple to host and reason about
- cheap to host
- comes pre configured for sub 15 dev teams
- completely open source, no paid ad-ons

It's not aimed at teams that can afford SREs, the idea was to provide a good tool for smaller teams and startups in the sub 15 dev range.

The base of Traceway is Clickhouse, nothing special there, if you want you can run it with sqlite for self hosting. Sessions are also stored in S3 so the costs are minimal.

It is opinionated, it comes with preconfigured SLOs for flagging issues with endpoints and it will never try to sell you an AI SRE, you can file your exceptions/slo issues with the git integration and run what ever AI you want on it (I was sick of observability tools trying to sell me an AI).

I'm a huge fan of open source, here is what we've done so far for making existing solutions better:

1 - Session Replays/RUM

Session replays are usually a premium/expensive feature. With Traceway you can self host them and add them to your app in minutes. I am working on making this a standalone feature that ties into the otel sdks for mobile/js so that you can get your spans/logs/metrics/exceptions from any platform connected to your session replays in Traceway. At one point I got nerd snipped into making it work with Flutter, so we are the only solution I know of that has affordable usable session replays for Flutter.

2 - Symfony Otel

Symfony, the php framework, had no library that offered a few line setup and worked out of the box with open telemetry. We wrote one, you can use it with any tool out there.

3 - Symbolicator

We're working on a symbolicator that will be Open Telemetry Collector compatible, so that you can get your stack traces for Js/Flutter/Android/iOS resolved back. From what I can tell no good solution exists for this currently.

## On the "per-language vendor SDK" question

A few people (rightly) called this out. To clarify: Traceway is fully OTel compliant.

**Go:** The original version started with Go SDKs. I've since moved to using Go OTel. I haven't updated those docs yet because the Go SDKs still work and are used in the wild, but they're on the deprecation track. Thanks for pointing it out.

**Symfony:** There were no good one-line OTel integrations out there for Symfony, so we wrote one. It is not a custom SDK, it's an OTel configurator. You can use it with any backend, not just Traceway. We're firm believers in contributing back to the OpenTelemetry community.

**Frontend / mobile:** This is more complicated. The current frontend and mobile OTel spec does not allow session replays to be sent, so for those platforms we still keep SDKs with a custom protocol alongside OTel. As soon as the spec matures I'm hoping to move it fully to OTel.

---

I hope that answers some of the questions you all have had and I am sorry I have not seen this earlier. I will make a proper HN post at some point with more info on the project, right now I am focusing on building. If you have any ideas or things you'd like to see feel free to comment, join our discord community or open the issue in our git, we're always happy to accept PRs.
