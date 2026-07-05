---
title: "ScyllaDB Is Faster Than Cassandra. So Why Is It #66?"
description: "What is ScyllaDB? Why Discord cut p99 reads from 40-125 ms to a steady 15 ms on 72 nodes, when the C++ Cassandra rewrite fits, and when MongoDB or SQL wins."
slug: "what-is-scylladb"
lang: "en"
alternate: "scylladb-la-gi.vi.md"
author: "ndmt1at21"
authorBio: "Backend engineer writing a devlog on distributed systems and databases."
date: "2026-07-05"
lastUpdated: "2026-07-05"
tags: ["scylladb", "database", "nosql", "cassandra", "mongodb", "distributed-systems"]
cover: "https://images.unsplash.com/photo-1695668548342-c0c1ad479aee?w=1200&h=630&fit=crop&q=80"
coverAlt: "Rows of rack-mounted servers glowing inside a modern data center, the kind of infrastructure a distributed database like ScyllaDB runs on."
coverGen: "A wide bank of glowing server racks linked by streams of data packets converging into one hexagonal database node, three-quarter view with the node at the right third, isometric flat vector illustration, dark navy background, cyan and orange accents, clean geometric lines, no gradients, 1200x630, no text, no words, no logos"
---

# ScyllaDB Is Faster Than Cassandra. So Why Is It #66?

_By **ndmt1at21**, backend engineer. Published 2026-07-05._

Here's a number that made a lot of engineers look twice. In 2023, Discord said it had replaced 177 Cassandra nodes with just 72 ScyllaDB nodes, and its p99 read latency on a store holding trillions of messages dropped from a jumpy 40-125 ms to a steady 15 ms ([Discord Engineering](https://discord.com/blog/how-discord-stores-trillions-of-messages)).

Cut your node count in half and your tail latency by 8x? That sounds like a free lunch. It isn't, and this post is about where the bill comes due.

I'll keep it practical. We'll cover what ScyllaDB actually is, the (genuinely fun) reason it exists, what its data model looks like in real CQL, when to reach for it, whether it can stand in for MongoDB or your SQL database, and the trade-offs the marketing pages skip. By the end you should be able to look at your own system and say "yes, this fits" or "no, walk away."

> **The short version**
>
> - ScyllaDB is a C++ rewrite of Apache Cassandra from the team behind the KVM hypervisor: same data model, no JVM, no garbage-collection pauses.
> - Discord's 2023 migration is the headline proof: p99 reads fell from 40-125 ms to a steady 15 ms.
> - It has no joins and no multi-partition transactions, so it replaces neither SQL nor MongoDB's query engine.
> - Most of the eye-popping "10x / 20x" numbers come from ScyllaDB's own benchmarks. Treat them as claims, not facts.

## What Is ScyllaDB?

ScyllaDB is a distributed NoSQL database that stores data in wide-column tables, the same shape Cassandra uses. Think of a giant, sorted key-value store spread across many machines: you pick a partition key, all the rows for that key live together, and you scale by adding nodes instead of sharding by hand in your app code.

The twist is under the hood. Cassandra runs on the JVM; ScyllaDB is written from scratch in C++ and released its first version in September 2015. It keeps Cassandra's tables, drivers, and CQL query language, then throws away the Java engine and replaces it with one designed for steady, predictable latency.

[IMAGE: Six hexagonal database nodes arranged in a ring, exchanging small data packets, illustrating a distributed wide-column cluster. | stock: https://images.unsplash.com/photo-1558494949-ef010cbdcc31?w=1200&h=800&fit=crop&q=80 | gen: Six hexagonal database nodes arranged in a ring, exchanging small data packets along every connecting line, centered ring floating in open space, isometric flat vector illustration, dark navy background, cyan and orange accents, clean geometric lines, no gradients, 3:2, no text, no words, no logos]

One nice practical detail: ScyllaDB also speaks Amazon DynamoDB's API through a layer called Alternator. So there are two doors in, from Cassandra or from DynamoDB, which matters if you're eyeing a migration.

And it's still shipping fast. Vector search went GA in early 2026, and the June 2026 release added DynamoDB Streams support and experimental strongly consistent tables. This is not a frozen project.

So why would anyone rewrite a perfectly good database in another language? That's the origin story, and it's a good one.

## Why Does ScyllaDB Exist?

Short answer: because a group of virtualization hackers looked at database overhead and thought "we've beaten worse."

In December 2014, Avi Kivity, the person who created the KVM hypervisor built into the Linux kernel, teamed up with Dor Laor and started the project at a company called Cloudius Systems. These weren't application developers who got frustrated with their ORM. Kivity had spent years making Linux itself run virtual machines efficiently.

The path there is half the fun. His team first built OSv, a stripped-down operating system for cloud workloads. OSv spun off something more important than itself: Seastar, an open-source C++ framework for building servers that share nothing between cores and never block. Seastar was good enough that other heavy hitters adopted it too, including Redpanda and Ceph's next-generation storage engine.

Then came the bet. Cassandra's distributed design is genuinely good, the developers reasoned, but its Java implementation pays a permanent tax: garbage-collection pauses that freeze the process, threads fighting over locks, I/O handed off to the kernel. What if you kept the proven design and rewrote the engine in C++ on Seastar?

[CALLOUT] The lineage in one line: KVM (run VMs on Linux) → OSv (shrink the OS) → Seastar (bypass the OS) → ScyllaDB (rewrite Cassandra on Seastar). Every step is the same move: delete a layer of overhead between your code and the metal.

That bet is the whole product. Keep the API developers already know, replace the part that stalls.

## How Does Shard-Per-Core Actually Work?

Here's the one idea that makes ScyllaDB tick, in plain terms. Instead of a pool of threads all sharing the node's data and fighting over locks, ScyllaDB gives each CPU core its own private slice of the data and its own to-do list. Cores don't share memory; they pass messages, like coworkers dropping notes in each other's inbox rather than all reaching into one drawer at once.

Picture a busy kitchen. The old model is ten cooks crowding one prep station, constantly bumping elbows. Shard-per-core gives each cook their own station, their own ingredients, their own tickets. Nobody waits on anybody. That's why throughput scales cleanly as you add cores.

[IMAGE: A stylized CPU with an isolated lane of data flowing into each core, illustrating shard-per-core pinning data and work to individual cores. | stock: https://images.unsplash.com/photo-1513366976578-e01c21fb9c76?w=1200&h=675&fit=crop&q=80 | gen: A stylized CPU chip with four glowing lanes of data packets flowing into four separate cores, each lane fully isolated from the others, chip centered on an uncluttered backdrop in an abstract diagram style, isometric flat vector illustration, dark navy background, cyan and orange accents, clean geometric lines, no gradients, 16:9, no text, no words, no logos]

The other half is the missing JVM. No Java means no stop-the-world garbage collection, which was historically the number-one cause of Cassandra's random latency spikes. ScyllaDB also runs its own I/O and compaction schedulers in userspace, so background cleanup doesn't stampede over live queries.

Does it actually pay off? ScyllaDB's own 2021 benchmark on identical hardware claims 2x to 5x Cassandra's throughput. That's a vendor number, so keep some salt handy. To their credit, the same report admits Cassandra slightly wins under low load, because ScyllaDB's schedulers keep doing small bits of background work when the system is idle. Honest benchmarks tell you where they lose, and that line is the tell.

## When Should You Use ScyllaDB?

Reach for ScyllaDB when three things are all true: you have serious write and read volume, your access patterns are known and simple (fetch by key, scan a partition), and you actually care about p99 latency, not just the average.

Discord is the textbook fit. By early 2022 it ran 177 Cassandra nodes holding trillions of messages, up from 12 nodes in 2017, and it was losing the fight against hot partitions and GC pauses. After moving to 72 ScyllaDB nodes in May 2022, p99 reads settled at 15 ms and p99 writes at 5 ms. Its custom Rust migrator pushed up to 3.2 million messages per second and finished the whole move in about nine days instead of a projected three months.

[CHART: Dumbbell chart, "Discord message-store p99 latency: Cassandra vs ScyllaDB (ms)". Data: reads 40-125 ms → 15 ms; writes 5-70 ms → 5 ms. Source: Discord Engineering, "How Discord Stores Trillions of Messages", 2023]

So what does "known and simple access patterns" look like in the wild? The sweet spot: messaging and chat history, activity and notification feeds, time-series and IoT telemetry, session or device state, and product catalogs. The common thread is write-heavy data you slice by a natural key and read along a handful of predictable paths.

A quick gut check before you commit. ScyllaDB probably fits if you can answer yes to most of these:

- Can you name your top five queries today, and will they still be the top five next year?
- Does every hot query start with a natural partition key (user, channel, device, tenant)?
- Are you writing a lot more than you're running ad-hoc reports?
- Do you have a real p99 target, like "reads under 20 ms," that you're currently missing?

Customer stories published by ScyllaDB echo the same shape, with the obvious caveat that the vendor is the one publishing them: Comcast's X1 platform went from 962 Cassandra nodes plus 60 cache servers down to 78 ScyllaDB nodes and reported around $2.5M saved per year, Rakuten went from 24 nodes to 6, and iFood reported roughly a 9x database cost cut. Great directional evidence, just remember who's telling the story.

## What Does a ScyllaDB Data Model Look Like?

This is where it gets concrete, and where the "free lunch" starts costing something. You don't design ScyllaDB tables around your data; you design them around your queries. Let's build the Discord-style example.

Say the query is "show the latest 50 messages in a channel." You'd model it roughly like this:

```sql
CREATE TABLE messages (
    channel_id  bigint,
    bucket      int,      -- a fixed time window, keeps each partition bounded
    message_id  bigint,   -- a time-sortable ID (Snowflake-style)
    author_id   bigint,
    content     text,
    PRIMARY KEY ((channel_id, bucket), message_id)
) WITH CLUSTERING ORDER BY (message_id DESC);
```

The partition key is `(channel_id, bucket)`, so every message in a channel-window lives together on the same nodes, already sorted newest-first. The read is boring and fast:

```sql
SELECT * FROM messages
WHERE channel_id = ? AND bucket = ?
LIMIT 50;
```

No scatter-gather, no cross-node coordination. One partition, sorted, done. That `bucket` column exists so a wildly popular channel doesn't grow one giant partition, a classic hot-partition trap.

[CALLOUT] The rule you can't forget: an efficient ScyllaDB query must include the partition key. No partition key means a full-cluster scan, which is the "why is production on fire" query.

Now the catch. Say product asks for a new screen: "all messages by a given user." Your `messages` table can't answer that, because it's partitioned by channel, not author. In SQL you'd add an index or a `WHERE author_id = ?`. In ScyllaDB you build a second table:

```sql
CREATE TABLE messages_by_author (
    author_id   bigint,
    message_id  bigint,
    channel_id  bigint,
    content     text,
    PRIMARY KEY (author_id, message_id)
) WITH CLUSTERING ORDER BY (message_id DESC);
```

And you write to both tables on every insert. That duplication is the price of query-first modeling: you trade storage and write complexity for dead-simple, fast reads. When your queries are stable, it's a great trade. When product invents a new filter every sprint, it's death by a thousand tables.

## Can ScyllaDB Replace MongoDB?

Sometimes yes, and Discord is the proof in both directions. Way back in November 2015, Discord actually left MongoDB at around 100 million messages because, in its own words, "the data and the index could no longer fit in RAM and latencies started to become unpredictable" ([Discord Engineering](https://discord.com/blog/how-discord-stores-billions-of-messages)). So the pattern where ScyllaDB wins is real.

[IMAGE: Three parallel highways of data packets racing toward a glowing database, a visual metaphor for high-throughput, low-latency writes. | stock: https://images.unsplash.com/photo-1741096931391-54a1717b44e0?w=1200&h=800&fit=crop&q=80 | gen: Three parallel highways of streaming data packets racing toward a glowing cylindrical database on the horizon, low rear perspective emphasizing motion lines, isometric flat vector illustration, dark navy background, cyan and orange accents, clean geometric lines, no gradients, 3:2, no text, no words, no logos]

If your MongoDB is really just a huge key-value or wide-column store with known access paths, ScyllaDB can take over, usually on fewer nodes. A 2023 benchANT study (sponsored by ScyllaDB, so read the label) reported ScyllaDB winning 132 of 133 YCSB measurements, with up to 20x higher throughput.

But read that study's own fine print, because it's the honest part. MongoDB won one measurement, MongoDB failed to even load the full 10 TB dataset, and aggregations and scans were never tested. Which workloads justify MongoDB's heavier query engine? Exactly the aggregations and scans the benchmark skipped. It's a bit like winning a race by only measuring the straightaways.

And MongoDB does two things ScyllaDB simply can't. It's had multi-document ACID transactions since 2018 and sharded distributed transactions since 2019. Its secondary indexes are flexible and ad-hoc. ScyllaDB's secondary indexes are backed by materialized views, one column each, with write overhead, nowhere near MongoDB's "just add a filter" ergonomics.

> **My rule of thumb:** the deciding factor isn't scale, it's query entropy. If you can freeze your query list, ScyllaDB fits. If your queries change constantly, MongoDB's flexible engine is precisely the thing you'd be throwing away.

## Can ScyllaDB Replace Your SQL Database?

Mostly, no, and you should be glad it's honest about that. ScyllaDB's own docs are blunt: it's a BASE system, not ACID. A CQL `BATCH` is atomic only within a single partition, lightweight transactions are single-partition compare-and-set only, and joins don't exist ([ScyllaDB Docs](https://docs.scylladb.com/manual/stable/kb/consistency.html)).

The real wall is modeling, and you already saw it above. Relational databases let you normalize now and invent queries later; the query planner figures out joins at runtime. ScyllaDB flips that: you enumerate queries first, build one denormalized table per query, and duplicate on write. There's no `JOIN users ON orders.user_id` waiting to bail you out.

[CALLOUT] SQL vs ScyllaDB in one sentence: SQL lets you store data now and ask anything later. ScyllaDB makes you decide the questions up front and store the answers.

[IMAGE: One dataset copied into three denormalized tables, each shaped to answer a single predefined query, illustrating query-first modeling. | stock: none | gen: One cube of raw data on the left splitting into three differently shaped denormalized tables, each table slotting into a matching query arrow like a puzzle piece, left-to-right abstract flow diagram, isometric flat vector illustration, dark navy background, cyan and orange accents, clean geometric lines, no gradients, 16:9, no text, no words, no logos]

There is movement at the edges: the June 2026 release ships experimental strongly consistent tables. "Experimental" is doing heavy lifting in that sentence, so don't put your billing ledger on it yet. If your domain needs cross-entity invariants, foreign keys, or free-form reporting queries, stay on Postgres or MySQL. That's not a knock on ScyllaDB; it's just a different tool.

## Try It in Five Minutes

You don't have to take any of this on faith. ScyllaDB ships an official Docker image, so you can poke at the data model on your laptop:

```bash
docker run --name scylla -d scylladb/scylla
# wait ~30s for it to boot, then open a CQL shell:
docker exec -it scylla cqlsh
```

From there, paste the `messages` table above, insert a few rows, and try a `SELECT` without the partition key. ScyllaDB will make you add `ALLOW FILTERING` and warn you, which is exactly the lesson: the database is telling you that query won't scale. Feeling that friction once teaches query-first modeling better than any blog post can.

## ScyllaDB Pros and Cons: An Honest Scorecard

Let's put it plainly: the engineering is impressive, the adoption is niche. In DB-Engines' July 2026 popularity ranking, ScyllaDB sits at #66 (score 4.94), while Cassandra is #10 (103.06) and MongoDB is #5 (386.62) ([DB-Engines](https://db-engines.com/en/ranking)). That gap is the honest starting point, and it's the answer to this post's title: faster on paper doesn't automatically mean popular.

**What's genuinely good:**

- Real, independent production proof: Discord's 72 nodes, trillions of messages, steady 15 ms p99 reads.
- Fewer machines for the same job, which is the whole cost argument.
- Two migration doors: Cassandra's CQL and DynamoDB's API.
- A fast, active roadmap: tablets, vector search, DynamoDB Streams, all shipped recently.

**What should give you pause:**

- Consistency has had sharp edges. In December 2020, [Jepsen's independent analysis](https://jepsen.io/analyses/scylla-4.2-rc3) of Scylla 4.2-rc3 found split-brain on healthy clusters and stale reads returning a value replaced 40+ seconds earlier. ScyllaDB fixed most of it and corrected docs that had overclaimed isolation, though Jepsen still flagged membership changes as risky and pushed for a Raft-based design, which ScyllaDB has since pursued.
- The license changed. On December 18, 2024, ScyllaDB dropped the AGPL for a source-available license; the free tier now caps at 10 TB and 50 vCPUs per organization. Some open-source folks were not happy about it.
- Small ecosystem. A #66 ranking means a smaller hiring pool, fewer ready-made integrations, and thinner answers when you're stuck at 3 a.m.
- Cassandra isn't sitting still. Version 5.0 added trie-indexed storage and a smarter compaction strategy, closing part of the gap you're paying to cross.

> **My take:** sort the evidence into two piles before you decide. The independent pile is Discord's production numbers, Jepsen's safety findings, and DB-Engines' adoption data. The vendor pile is every 10x, 20x, and 68x multiplier. Both are useful, but only the first pile should set your expectations.

[CHART: Grouped horizontal bar, "Node counts before and after migrating to ScyllaDB". Data: Discord 177 → 72 (independent, Discord blog); Comcast 962 + 60 cache servers → 78 (vendor-published); Rakuten 24 → 6 (vendor-published). Caption must label Comcast and Rakuten as ScyllaDB-published customer figures. Source: Discord Engineering, 2023; ScyllaDB customer talks, 2020-2022]

## Frequently Asked Questions

### Is ScyllaDB compatible with Apache Cassandra?

Mostly yes, by design. It launched in 2015 as a drop-in Cassandra replacement: same CQL, same drivers, same wide-column model. Still, test before you migrate, because internals like compaction and scheduling behave differently under real load, and that's where surprises hide.

### Is ScyllaDB still open source?

No, not since December 18, 2024, when it moved from the AGPL to a source-available license. OSS 6.2.x was the last AGPL release. The free tier now allows up to 10 TB of storage and 50 vCPUs per organization, which is plenty for a prototype but a real constraint at scale.

### How much faster is ScyllaDB than Cassandra, really?

The trustworthy answer is Discord's: p99 reads dropped from 40-125 ms to a steady 15 ms in production. ScyllaDB's own 2021 benchmark claims 2x to 5x throughput on the same hardware, but even it admits Cassandra wins under low load. Treat every vendor multiplier as a claim to verify, not a promise.

### Does ScyllaDB support ACID transactions?

Not in the way you'd mean it in SQL. A `BATCH` is atomic only within a single partition, and lightweight transactions do single-partition compare-and-set. There are no cross-partition, multi-row transactions. MongoDB, for contrast, has offered multi-document ACID transactions since 2018.

### Who actually runs ScyllaDB in production?

Discord is the best-documented user, with 72 nodes holding trillions of messages as of 2023. ScyllaDB-published customer talks add Comcast (962 Cassandra nodes down to 78), Rakuten (24 to 6), and iFood (about a 9x cost cut), though those numbers come from the vendor rather than an independent source.

## Conclusion

ScyllaDB is what you get when kernel-level engineers rewrite Cassandra in C++: shard-per-core, no GC pauses, and a genuinely independent win in Discord's drop from 40-125 ms to a steady 15 ms. It's a sharp specialist tool, not a general-purpose database.

- Reach for it when you have high-throughput, known-query, p99-sensitive workloads.
- Walk away if you need joins, multi-partition transactions, or ad-hoc analytics.
- Factor in the December 2024 license change and the small ecosystem before you commit.

The fastest way to decide? Write down your query list. Model those queries as tables like we did above. If that feels natural, ScyllaDB fits. If you keep wishing for a `JOIN`, you have your answer.

<!-- NOTE TO AUTHOR: No first-hand production stories are included because there's no ScyllaDB/Cassandra/MongoDB production experience on record for this author. If you've run any of these at scale (metrics, incidents, tuning wins), send them and I'll fold them in as a real first-hand section. -->

## Sources

- Discord Engineering (Bo Ingram), "How Discord Stores Trillions of Messages", retrieved 2026-07-05, https://discord.com/blog/how-discord-stores-trillions-of-messages
- Discord Engineering, "How Discord Stores Billions of Messages", retrieved 2026-07-05, https://discord.com/blog/how-discord-stores-billions-of-messages
- InfoQ, "Discord Migrates Trillions of Messages from Cassandra to ScyllaDB", retrieved 2026-07-05, https://www.infoq.com/news/2023/06/discord-cassandra-scylladb/
- Jepsen (Kyle Kingsbury), "Scylla 4.2-rc3", retrieved 2026-07-05, https://jepsen.io/analyses/scylla-4.2-rc3
- DB-Engines, "DB-Engines Ranking", retrieved 2026-07-05, https://db-engines.com/en/ranking
- MongoDB, "MongoDB Multi-Document ACID Transactions: General Availability", retrieved 2026-07-05, https://www.mongodb.com/company/blog/product-release-announcements/mongodb-multi-document-acid-transactions-general-availability
- ScyllaDB, "Apache Cassandra 4.0 vs. ScyllaDB 4.4: Comparing Performance", retrieved 2026-07-05, https://www.scylladb.com/2021/08/24/apache-cassandra-4-0-vs-scylla-4-4-comparing-performance/
- ScyllaDB, "Benchmarking Apache Cassandra (40 Nodes) vs ScyllaDB (4 Nodes)", retrieved 2026-07-05, https://www.scylladb.com/2022/05/18/benchmarking-apache-cassandra-40-nodes-vs-scylladb-4-nodes/
- benchANT, "NoSQL Benchmark: MongoDB vs ScyllaDB", retrieved 2026-07-05, https://benchant.com/blog/mongodb-vs-scylladb-benchmark
- ScyllaDB, "Cutting Database Costs: Lessons from Comcast, Rakuten & iFood", retrieved 2026-07-05, https://www.scylladb.com/2022/12/15/cutting-database-costs/
- ScyllaDB, "Scaling Performance Comparison: ScyllaDB Tablets vs Cassandra vNodes", retrieved 2026-07-05, https://www.scylladb.com/2026/01/13/scaling-performance-comparison-vs-cassandra/
- ScyllaDB, "Why We're Moving to a Source Available License", retrieved 2026-07-05, https://www.scylladb.com/2024/12/18/why-were-moving-to-a-source-available-license/
- Peter Zaitsev, "Thoughts on ScyllaDB License Change", retrieved 2026-07-05, https://peterzaitsev.com/thoughts-on-scylladb-license-change/
- ScyllaDB Docs, "Consistency in ScyllaDB", retrieved 2026-07-05, https://docs.scylladb.com/manual/stable/kb/consistency.html
- ScyllaDB Docs, "Global Secondary Indexes", retrieved 2026-07-05, https://docs.scylladb.com/manual/stable/features/secondary-indexes.html
- ScyllaDB, "The ScyllaDB Story", retrieved 2026-07-05, https://www.scylladb.com/company/the-scylla-story/
- ScyllaDB, "ScyllaDB 2026.2", retrieved 2026-07-05, https://www.scylladb.com/2026/06/29/scylladb-2026-2/
- Wikipedia, "ScyllaDB", retrieved 2026-07-05, https://en.wikipedia.org/wiki/ScyllaDB
- Wikipedia, "Avi Kivity", retrieved 2026-07-05, https://en.wikipedia.org/wiki/Avi_Kivity
- CMU Database of Databases (dbdb.io), "Scylla", retrieved 2026-07-05, https://dbdb.io/db/scylla
- Instaclustr, "Exploring the key features of Cassandra 5.0", retrieved 2026-07-05, https://www.instaclustr.com/blog/exploring-the-key-features-of-cassandra-5-0/
