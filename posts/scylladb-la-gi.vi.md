---
title: "ScyllaDB nhanh hơn Cassandra, sao vẫn đứng hạng 66?"
description: "ScyllaDB là gì? Vì sao Discord giảm p99 read từ 40-125 ms xuống 15 ms với 72 node, khi nào nên dùng bản C++ viết lại Cassandra, khi nào MongoDB hay SQL thắng."
slug: "scylladb-la-gi"
lang: "vi"
alternate: "what-is-scylladb.en.md"
author: "ndmt1at21"
authorBio: "Backend engineer, viết devlog về distributed systems và database."
date: "2026-07-05"
lastUpdated: "2026-07-05"
tags: ["scylladb", "database", "nosql", "cassandra", "mongodb", "distributed-systems"]
cover: "https://images.unsplash.com/photo-1695668548342-c0c1ad479aee?w=1200&h=630&fit=crop&q=80"
coverAlt: "Những dãy server gắn rack sáng đèn trong một data center hiện đại, đúng kiểu hạ tầng mà một distributed database như ScyllaDB vận hành trên đó."
coverGen: "A wide bank of glowing server racks linked by streams of data packets converging into one hexagonal database node, three-quarter view with the node at the right third, isometric flat vector illustration, dark navy background, cyan and orange accents, clean geometric lines, no gradients, 1200x630, no text, no words, no logos"
---

# ScyllaDB nhanh hơn Cassandra, sao vẫn đứng hạng 66?

_Tác giả **ndmt1at21**, backend engineer. Đăng ngày 05/07/2026._

Năm 2023, Discord công bố những con số đứng sau một trong những cuộc migration database được bàn tán nhiều nhất mấy năm gần đây: 177 node Cassandra rút xuống còn 72 node ScyllaDB, và p99 read latency trên kho dữ liệu chứa hàng nghìn tỷ tin nhắn giảm từ 40-125 ms xuống mức ổn định 15 ms ([Discord Engineering, "How Discord Stores Trillions of Messages"](https://discord.com/blog/how-discord-stores-trillions-of-messages)). Những con số như vậy lan rất nhanh. Và chúng đặt ra một câu hỏi chính đáng: rốt cuộc ScyllaDB là gì, và tại sao không phải ai cũng dùng?

Bài này là notes nghiên cứu của mình, viết lại thành một hướng dẫn cho backend engineer đang cân nhắc chọn database. Chúng ta sẽ đi qua: ScyllaDB từ đâu ra, kiến trúc shard-per-core hoạt động thế nào, workload nào thực sự hợp, nó có thay được MongoDB hay database SQL của bạn không, và cả những điểm trừ mà vendor sẽ không tự nói ra. [INTERNAL-LINK: chọn database cho hệ thống throughput cao → bài pillar về lựa chọn database]

> **Tóm tắt nhanh**
>
> - Bản viết lại Apache Cassandra bằng C++ từ team của Avi Kivity, cha đẻ KVM: không JVM, không GC pause.
> - Năm 2023, Discord báo cáo p99 read giảm từ 40-125 ms xuống 15 ms sau khi migrate (Discord Engineering).
> - Không JOIN, không transaction đa partition: nó không thay được SQL hay query engine của MongoDB.
> - Đa số con số nhân "x lần" đến từ benchmark của vendor; hãy đọc chúng như lời quảng cáo có điều kiện.

## ScyllaDB là gì?

ScyllaDB là một distributed database dạng wide-column NoSQL, được viết lại từ đầu bằng C++ để thay thế trực tiếp (drop-in) cho Apache Cassandra, phát hành lần đầu ngày 22/09/2015 (Wikipedia, "ScyllaDB"). Nó giữ nguyên data model, driver và ngôn ngữ truy vấn CQL của Cassandra, nhưng thay JVM bằng một engine shard-per-core được thiết kế cho tail latency ổn định.

[IMAGE: Những bó cáp mạng được đi gọn gàng nối các thiết bị trong phòng server, minh họa tầng networking của một cluster database phân tán (Ảnh: Taylor Vick, Unsplash). | stock: https://images.unsplash.com/photo-1558494949-ef010cbdcc31?w=1200&h=800&fit=crop&q=80 | gen: Six hexagonal database nodes arranged in a ring, exchanging small data packets along every connecting line, centered ring floating in open space, isometric flat vector illustration, dark navy background, cyan and orange accents, clean geometric lines, no gradients, 3:2, no text, no words, no logos]

"Wide-column" nghĩa là dữ liệu nằm trong các bảng được chia theo partition key, các row được sắp xếp sẵn bên trong từng partition. Bạn scale bằng cách thêm node, chứ không phải tự shard trong code ứng dụng. ScyllaDB còn nói được API của Amazon DynamoDB qua một lớp tên là Alternator, tức là nó nhắm cùng lúc hai đường migration.

Và nó vẫn đang tiến rất nhanh. Vector search đạt GA vào tháng 1/2026, còn bản 2026.2 phát hành ngày 29/06/2026 bổ sung tương thích DynamoDB Streams cùng strongly consistent tables ở dạng thử nghiệm (ScyllaDB, "ScyllaDB 2026.2").

> ScyllaDB là một distributed database dạng wide-column, giấy phép source-available, phát hành lần đầu năm 2015 dưới dạng bản viết lại hoàn toàn bằng C++ của Apache Cassandra. Nó tương thích cả CQL của Cassandra lẫn API DynamoDB của Amazon; bản 2026.2 thêm DynamoDB Streams và strongly consistent tables thử nghiệm (ScyllaDB, 2026).

[INTERNAL-LINK: các data model NoSQL cơ bản → bài giới thiệu về key-value, document và wide-column store]

Vậy tại sao lại có người đi viết lại một database đang chạy tốt bằng ngôn ngữ khác? Đó chính là câu chuyện khởi nguồn, và nó khá thú vị.

## Vì sao ScyllaDB ra đời?

ScyllaDB ra đời vì những kỹ sư kỳ cựu về ảo hóa coi overhead của database là một bài toán hệ thống. Tháng 12/2014, Avi Kivity, người tạo ra hypervisor KVM, cùng Dor Laor khởi động dự án tại Cloudius Systems và tung bản open source đầu tiên vào tháng 9/2015 (ScyllaDB, "The ScyllaDB Story"; Wikipedia, "Avi Kivity").

Kivity là người xây KVM, máy ảo trong nhân Linux, bắt đầu từ năm 2006 tại Qumranet, startup được Red Hat mua lại năm 2008. Sau đó team của ông tại Cloudius xây OSv, một hệ điều hành unikernel cho cloud. OSv sinh ra một sản phẩm phụ còn quan trọng hơn: Seastar, framework C++ giấy phép Apache 2.0 cho các server bất đồng bộ, shared-nothing. Ngày nay Seastar còn là nền của Redpanda và dự án Crimson của Ceph (dbdb.io, "Scylla").

Rồi đến ván cược. Thiết kế phân tán của Cassandra là đúng đắn, nhưng phần hiện thực bằng Java phải trả một khoản thuế thường trực: GC pause, tranh chấp thread, I/O do kernel quản lý. Nếu giữ nguyên thiết kế nhưng viết lại engine bằng C++ trên Seastar thì sao? Câu hỏi đó trở thành ScyllaDB.

[CALLOUT] Dòng dõi gói trong một dòng: KVM (ảo hóa Linux) → OSv (thu nhỏ OS) → Seastar (đi vòng qua OS) → ScyllaDB (viết lại Cassandra trên Seastar). Mỗi bước đều nhắm cùng một kẻ thù: các lớp overhead giữa code của bạn và phần cứng.

> ScyllaDB được thành lập tháng 12/2014 bởi Avi Kivity và Dor Laor tại Cloudius Systems, công ty đứng sau unikernel OSv và framework C++ Seastar. Trước đó Kivity tạo ra KVM, hypervisor trong nhân Linux, tại Qumranet, công ty được Red Hat mua lại năm 2008 (Wikipedia; dbdb.io).

## Bên trong kiến trúc shard-per-core

Shard-per-core nghĩa là mỗi CPU core sở hữu riêng một phần dữ liệu của node và chạy không lock, không shared memory. Benchmark năm 2021 do chính ScyllaDB thực hiện trên cùng phần cứng tuyên bố thiết kế này đạt throughput gấp 2 đến 5 lần Apache Cassandra 4.0 (ScyllaDB, "Apache Cassandra 4.0 vs. ScyllaDB 4.4: Comparing Performance"). Con số của vendor, nên hãy khoan tin vội.

[IMAGE: Một CPU cách điệu với làn dữ liệu riêng chảy vào từng core, minh họa kiến trúc shard-per-core của ScyllaDB, nơi dữ liệu và công việc được ghim vào từng core (phương án stock, Ảnh: Alexandru-Bogdan Ghita, Unsplash). | stock: https://images.unsplash.com/photo-1513366976578-e01c21fb9c76?w=1200&h=675&fit=crop&q=80 | gen: A stylized CPU chip with four glowing lanes of data packets flowing into four separate cores, each lane fully isolated from the others, chip centered on an uncluttered backdrop in an abstract diagram style, isometric flat vector illustration, dark navy background, cyan and orange accents, clean geometric lines, no gradients, 16:9, no text, no words, no logos]

Bên dưới, Seastar cấp cho mỗi core một shard dữ liệu riêng, bộ nhớ riêng (NUMA-aware) và hàng đợi task riêng. Các core trao đổi bằng message passing thay vì tranh nhau lock. I/O chạy qua các scheduler userspace của chính ScyllaDB thay vì dựa vào kernel, và compaction cũng bị chính các scheduler đó điều tiết. Không có JVM nghĩa là không có GC stop-the-world, thứ xưa nay là nguồn lớn nhất gây spike tail latency cho Cassandra.

Công bằng mà nói, chính báo cáo 2021 đó có một lời thú nhận hiếm hoi: "Under low load Cassandra slightly outperforms ScyllaDB" (tải thấp thì Cassandra nhỉnh hơn), một phần do compaction chạy lúc rảnh và scheduler tick 0,5 ms. Một bài test khác cũng do ScyllaDB tự chạy năm 2022 tuyên bố 4 node ScyllaDB thay được 40 node Cassandra với throughput cao hơn 42% và chi phí thấp hơn khoảng 2,5 lần (ScyllaDB, "Benchmarking Apache Cassandra (40 Nodes) vs ScyllaDB (4 Nodes)"). Ấn tượng nếu đúng, nhưng người chạy đua và trọng tài là cùng một công ty.

> ScyllaDB gán một shard dữ liệu và công việc cho mỗi CPU core. Các core giao tiếp bằng message passing thay vì lock trên shared memory, cấp phát bộ nhớ theo NUMA, còn I/O chạy qua scheduler ở userspace. Không có JVM nên không có GC pause, nguyên nhân kinh điển của spike tail latency trên Cassandra (ScyllaDB GitHub; dbdb.io).

[INTERNAL-LINK: hiểu p99 và tail latency → bài về latency percentile và SLO]

## Khi nào nên dùng ScyllaDB?

Dùng ScyllaDB khi bạn cần throughput cao liên tục với p99 latency dự đoán được, trên các mẫu truy cập kiểu key-value hoặc wide-column. Năm 2023, Discord báo cáo đúng hồ sơ đó: p99 read giảm từ 40-125 ms trên Cassandra xuống ổn định 15 ms trên ScyllaDB, p99 write từ 5-70 ms xuống 5 ms (Discord Engineering, "How Discord Stores Trillions of Messages").

[CHART: Biểu đồ dumbbell, "p99 latency kho tin nhắn Discord: Cassandra vs ScyllaDB (ms)". Dữ liệu: read 40-125 ms → 15 ms; write 5-70 ms → 5 ms. Nguồn: Discord Engineering, "How Discord Stores Trillions of Messages", 2023]

Chi tiết khiến câu chuyện cụ thể hơn. Đầu năm 2022, Discord chạy 177 node Cassandra chứa hàng nghìn tỷ tin nhắn, tăng từ 12 node năm 2017, trong khi vật lộn với hot partition và GC pause. Họ chuyển hẳn sang 72 node ScyllaDB vào tháng 5/2022, mỗi node mang 9 TB disk so với trung bình khoảng 4 TB trước đó. Migrator tự viết bằng Rust của họ đẩy tới 3,2 triệu tin nhắn mỗi giây và hoàn tất trong khoảng chín ngày thay vì ba tháng như dự tính (InfoQ, "Discord Migrates Trillions of Messages from Cassandra to ScyllaDB", 2023).

Workload của bạn có thực sự giống vậy không? Danh sách hợp gu: tin nhắn và lịch sử chat, activity feed, time series và telemetry IoT, trạng thái session hay thiết bị, catalog. Nói gọn: dữ liệu nặng về ghi, chia partition được theo một khóa tự nhiên, truy vấn theo những đường đã biết trước, và có SLO p99 thật sự.

Các bài talk khách hàng do ScyllaDB đăng năm 2022 lặp lại đúng mẫu hình đó, kèm lưu ý hiển nhiên rằng chính vendor là bên công bố: nền tảng X1 của Comcast, phục vụ hơn 30 triệu set-top box với khoảng 2 tỷ REST call mỗi ngày, đi từ 962 node Cassandra cộng 60 server cache xuống 78 node ScyllaDB, báo cáo tiết kiệm khoảng 2,5 triệu USD mỗi năm; Rakuten từ 24 node xuống 6; iFood báo cáo giảm chi phí database khoảng 9 lần (ScyllaDB, "Cutting Database Costs: Lessons from Comcast, Rakuten & iFood").

> Năm 2023, Discord báo cáo việc chuyển kho tin nhắn từ 177 node Cassandra sang 72 node ScyllaDB đã giảm p99 read latency từ 40-125 ms xuống ổn định 15 ms, còn p99 write giảm từ 5-70 ms xuống 5 ms (Discord Engineering, "How Discord Stores Trillions of Messages", 2023). Đó chính xác là loại workload mà ScyllaDB sinh ra để phục vụ.

[INTERNAL-LINK: thiết kế partition key tránh hot spot → bài về data modeling cho hệ thống phân tán]

## ScyllaDB có thay thế được MongoDB không?

Có lúc được, và Discord là bằng chứng theo cả hai chiều. Tháng 11/2015, Discord rời MongoDB ở mốc khoảng 100 triệu tin nhắn vì, theo đúng lời họ, "the data and the index could no longer fit in RAM and latencies started to become unpredictable" (dữ liệu và index không còn vừa RAM, latency bắt đầu khó lường) ([Discord Engineering, "How Discord Stores Billions of Messages"](https://discord.com/blog/how-discord-stores-billions-of-messages), 2017).

[IMAGE: Vệt sáng phơi sáng dài của dòng xe ban đêm trên cao tốc, ẩn dụ thị giác cho luồng dữ liệu throughput cao, độ trễ thấp (Ảnh: Egor Litvinov, Unsplash). | stock: https://images.unsplash.com/photo-1741096931391-54a1717b44e0?w=1200&h=800&fit=crop&q=80 | gen: Three parallel highways of streaming data packets racing toward a glowing cylindrical database on the horizon, low rear perspective emphasizing motion lines, isometric flat vector illustration, dark navy background, cyan and orange accents, clean geometric lines, no gradients, 3:2, no text, no words, no logos]

Nếu hệ MongoDB của bạn thực chất là một kho key-value hay wide-column khổng lồ với các đường truy cập đã biết trước, ScyllaDB hoàn toàn thay được, thường với ít node hơn. Năm 2023, [nghiên cứu benchANT "NoSQL Benchmark: MongoDB vs ScyllaDB"](https://benchant.com/blog/mongodb-vs-scylladb-benchmark) do ScyllaDB tài trợ báo cáo ScyllaDB thắng 132 trên 133 phép đo YCSB, throughput cao hơn tới 20 lần và price/performance tốt hơn tới 19 lần.

Nhưng hãy đọc phần chữ nhỏ của chính nghiên cứu đó trước khi trích dẫn. MongoDB thắng một phép đo, YCSB có vấn đề coordinated omission đã biết, MongoDB không load nổi trọn bộ dữ liệu 10 TB, còn aggregation và scan thì hoàn toàn không được test. Những workload nào khiến query engine của MongoDB xứng đáng với overhead của nó? Chính xác là những thứ benchmark đã bỏ qua.

MongoDB cũng đã hỗ trợ multi-document ACID transaction từ v4.0 năm 2018, và distributed transaction trên sharded cluster từ v4.2 năm 2019 (MongoDB, "MongoDB Multi-Document ACID Transactions: General Availability"). ScyllaDB không có cả hai. Secondary index của nó dựng trên materialized view, mỗi index một cột, kèm chi phí ghi: còn rất xa mới chạm tới khả năng index ad-hoc và aggregation pipeline của MongoDB.

<!-- [UNIQUE INSIGHT] -->
> **Góc nhìn của mình:** yếu tố quyết định không phải quy mô, mà là độ "hỗn loạn" của truy vấn (query entropy). Nếu hôm nay bạn viết ra được danh sách truy vấn chính và sang năm chúng vẫn y nguyên, ScyllaDB hợp. Còn nếu team product mỗi sprint lại xin thêm một filter hay một phép aggregation mới, thì query engine của MongoDB chính là thứ bạn sắp vứt đi.

> ScyllaDB có thể thay MongoDB cho các workload key-value và wide-column ở quy mô lớn, như kho tin nhắn của Discord đã chứng minh sau khi MongoDB chạm trần RAM ở mốc khoảng 100 triệu tin nhắn năm 2015. Nó không thể thay MongoDB cho truy vấn ad-hoc, aggregation, hay multi-document ACID transaction, thứ MongoDB đã có từ 2018 (Discord Engineering; MongoDB).

[INTERNAL-LINK: các pattern thiết kế schema MongoDB → bài về trade-off khi modeling document]

## ScyllaDB có thay thế được database quan hệ (SQL) không?

Phần lớn là không. Tính đến 2026, tài liệu chính thức của ScyllaDB nói thẳng: đây là hệ BASE, không phải ACID. Một CQL BATCH chỉ atomic và isolated bên trong một partition duy nhất, lightweight transaction (LWT) chỉ là compare-and-set trong một partition, còn JOIN thì đơn giản là không tồn tại ([ScyllaDB Docs, "Consistency in ScyllaDB"](https://docs.scylladb.com/manual/stable/kb/consistency.html)).

Độ vênh sâu hơn nằm ở cách modeling. Database quan hệ cho bạn chuẩn hóa dữ liệu trước rồi nghĩ ra truy vấn sau. ScyllaDB đảo ngược điều đó: bạn liệt kê truy vấn trước, thiết kế một bảng phi chuẩn hóa cho mỗi đường truy cập, và mọi truy vấn hiệu quả đều phải kèm partition key (ScyllaDB Docs, "Global Secondary Indexes"). Bạn có liệt kê nổi mọi truy vấn mà ứng dụng của mình sẽ chạy trong tương lai không? Với sản phẩm thiên về analytics, không ai làm nổi.

[CALLOUT] Query-first modeling gói trong một câu: SQL cho bạn lưu dữ liệu trước rồi muốn hỏi gì thì hỏi; ScyllaDB bắt bạn chốt câu hỏi từ đầu và lưu sẵn câu trả lời.

[IMAGE: Minh họa dạng sơ đồ cho query-first data modeling: một nguồn dữ liệu được nhân bản thành ba bảng phi chuẩn hóa, mỗi bảng sinh ra để trả lời một truy vấn định trước. | stock: none | gen: One cube of raw data on the left splitting into three differently shaped denormalized tables, each table slotting into a matching query arrow like a puzzle piece, left-to-right abstract flow diagram, isometric flat vector illustration, dark navy background, cyan and orange accents, clean geometric lines, no gradients, 16:9, no text, no words, no logos]

Ở rìa thì có chuyển động. ScyllaDB 2026.2, phát hành ngày 29/06/2026, mang theo Strongly Consistent Tables thử nghiệm và migration online từ vNode sang tablet (ScyllaDB, "ScyllaDB 2026.2"). "Thử nghiệm" là từ khóa chịu lực ở đây: đừng đặt cược sổ cái kế toán vào nó. Nếu domain của bạn cần ràng buộc xuyên thực thể, foreign key, hay truy vấn báo cáo, hãy ở lại với hệ quan hệ.

> ScyllaDB không có JOIN và không có ACID transaction đa partition. Câu lệnh CQL BATCH chỉ atomic và isolated trong một partition, còn lightweight transaction là phép compare-and-set kiểu Paxos trong một partition duy nhất (ScyllaDB Docs, "Consistency in ScyllaDB"). Thay một schema quan hệ đồng nghĩa với thiết kế lại quanh truy vấn, không phải migrate là xong.

[INTERNAL-LINK: khi nào phi chuẩn hóa có lời → bài về trade-off chuẩn hóa trong hệ OLTP]

## Ưu và nhược điểm của ScyllaDB: đánh giá thẳng thắn

Kỹ thuật của ScyllaDB rất ấn tượng; độ phổ biến thì vẫn là thị trường ngách. Trong bảng xếp hạng DB-Engines tháng 7/2026, ScyllaDB đứng thứ 66 với 4,94 điểm, so với Cassandra hạng 10 (103,06) và MongoDB hạng 5 (386,62) ([DB-Engines, "DB-Engines Ranking"](https://db-engines.com/en/ranking)). Mọi đánh giá trung thực nên bắt đầu từ khoảng cách đó.

**Ưu điểm:**

- Bằng chứng production độc lập: con số 2023 của Discord, 72 node, hàng nghìn tỷ tin nhắn, p99 read ổn định 15 ms (Discord Engineering).
- Ít node hơn cho cùng khối lượng việc: 4 so với 40 trong benchmark 2022 của chính ScyllaDB, 962 xuống 78 trong bài talk Comcast do vendor đăng.
- Hai lối thoát tương thích: CQL của Cassandra và API DynamoDB (Alternator).
- Tốc độ phát triển nhanh: bản 2025.4 (tháng 1/2026) mở rộng tablets sang index, CDC và LWT; một benchmark vendor tháng 1/2026 tuyên bố mở rộng cluster nhanh hơn 7,2 lần so với vNodes của Cassandra 5.0 (ScyllaDB, "Scaling Performance Comparison: ScyllaDB Tablets vs Cassandra vNodes").

**Nhược điểm:**

- Những góc sắc về consistency là có thật. Tháng 12/2020, [phân tích độc lập của Jepsen trên Scylla 4.2-rc3](https://jepsen.io/analyses/scylla-4.2-rc3) phát hiện split-brain trên cluster khỏe mạnh, stale read trả về giá trị đã bị thay từ hơn 40 giây trước, và vi phạm isolation ở các batch update không dùng LWT ngay cả tại consistency level ALL. ScyllaDB sửa phần lớn lỗi trong 4.1.9, 4.2.1 và 4.3-rc1, đồng thời sửa lại tài liệu từng nói quá về isolation. Jepsen vẫn lưu ý rằng các thay đổi membership và topology nhiều khả năng vẫn còn vấn đề, và khuyến nghị chuyển sang hệ thống membership dựa trên Raft, hướng mà ScyllaDB đã theo đuổi sau đó.
- Rủi ro giấy phép: ngày 18/12/2024, ScyllaDB bỏ AGPL để chuyển sang giấy phép source-available. OSS 6.2.x là bản open source cuối cùng, và bậc miễn phí giới hạn 10 TB lưu trữ cùng 50 vCPU cho mỗi tổ chức (ScyllaDB, "Why We're Moving to a Source Available License"). Những người thực chiến như Peter Zaitsev đã chỉ trích bước đi này (Peter Zaitsev, "Thoughts on ScyllaDB License Change").
- Lực hút hệ sinh thái: hạng 66 đồng nghĩa nguồn tuyển dụng nhỏ hơn, ít integration được tôi luyện hơn, và ít câu trả lời từ cộng đồng hơn khi bạn debug lúc 3 giờ sáng.
- Cassandra không đứng yên: bản 5.0 thêm trie memtables, SSTable đánh index kiểu trie và Unified Compaction Strategy (Instaclustr, "Exploring the key features of Cassandra 5.0"), thu hẹp một phần khoảng cách.

<!-- [UNIQUE INSIGHT] -->
> **Góc nhìn của mình:** trước khi quyết, hãy chia bằng chứng thành hai chồng. Chồng độc lập gồm số liệu production của Discord, phát hiện an toàn của Jepsen và dữ liệu phổ biến của DB-Engines. Chồng vendor gồm mọi con số nhân 10x, 20x, 68x. Cả hai chồng đều hữu ích; nhưng chỉ chồng thứ nhất nên định hình kỳ vọng của bạn.

[CHART: Biểu đồ cột ngang nhóm, "Số node trước và sau khi chuyển sang ScyllaDB". Dữ liệu: Discord 177 → 72 (độc lập, blog Discord); Comcast 962 + 60 server cache → 78 (vendor công bố); Rakuten 24 → 6 (vendor công bố). Chú thích phải ghi rõ số liệu Comcast và Rakuten là do ScyllaDB công bố. Nguồn: Discord Engineering, 2023; các bài talk khách hàng của ScyllaDB, 2020-2022]

> Trong bảng xếp hạng DB-Engines tháng 7/2026, ScyllaDB đứng thứ 66 với 4,94 điểm, trong khi Cassandra hạng 10 (103,06) và MongoDB hạng 5 (386,62) (DB-Engines, truy cập 2026-07-05). Câu chuyện hiệu năng thuần của ScyllaDB vẫn chưa chuyển hóa thành mức phổ biến đại chúng, nên quy mô hệ sinh thái vẫn là một chi phí thật.

## Câu hỏi thường gặp

### ScyllaDB có tương thích với Apache Cassandra không?

Có, ngay từ thiết kế. ScyllaDB ra mắt tháng 9/2015 như một bản thay thế drop-in cho Cassandra: nói CQL, chạy với driver Cassandra sẵn có, và giữ nguyên data model wide-column (Wikipedia, "ScyllaDB"). Dù vậy vẫn phải test trước khi migrate, vì những phần bên trong như compaction, scheduling và tuning hành xử khác nhau dưới tải.

### ScyllaDB còn là open source không?

Không. Ngày 18/12/2024, ScyllaDB chuyển từ AGPL sang giấy phép source-available, với OSS 6.2.x là bản AGPL cuối cùng. Bậc miễn phí cho phép tối đa 10 TB lưu trữ tổng và 50 vCPU cho mỗi tổ chức (ScyllaDB, "Why We're Moving to a Source Available License").

### ScyllaDB thực sự nhanh hơn Cassandra bao nhiêu?

Bằng chứng độc lập: cuộc migration 2023 của Discord giảm p99 read từ 40-125 ms xuống 15 ms. Benchmark 2021 của chính ScyllaDB tuyên bố throughput gấp 2 đến 5 lần trên cùng phần cứng, đồng thời thừa nhận Cassandra nhỉnh hơn khi tải thấp (ScyllaDB; Discord Engineering). Hãy coi mọi con số nhân của vendor là lời tuyên bố.

### ScyllaDB có hỗ trợ ACID transaction không?

Không theo nghĩa tổng quát. Tính đến 2026, CQL BATCH chỉ atomic trong một partition, còn LWT chỉ là compare-and-set trong một partition (ScyllaDB Docs, "Consistency in ScyllaDB"). Ngược lại, MongoDB đã có multi-document ACID transaction từ v4.0 năm 2018.

### Ai đang chạy ScyllaDB trong production?

Discord là ca được ghi chép kỹ nhất: 72 node chứa hàng nghìn tỷ tin nhắn tính đến 2023 (Discord Engineering). Các bài talk khách hàng do ScyllaDB đăng bổ sung Comcast (từ 962 node Cassandra xuống 78), Rakuten (24 xuống 6) và iFood (giảm chi phí database khoảng 9 lần); đó là các số liệu do vendor công bố.

[INTERNAL-LINK: các case study kiến trúc quy mô lớn → trang tag các bài distributed-systems]

## Kết luận

ScyllaDB là kết quả khi những kỹ sư tầng kernel viết lại Cassandra bằng C++: shard-per-core, không GC pause, và những chiến thắng được xác nhận độc lập như cú giảm p99 read 40-125 ms xuống 15 ms của Discord năm 2023. Nhưng nó là một chuyên gia, không phải kho dữ liệu vạn năng.

- Chọn nó cho workload throughput cao, truy vấn biết trước, nhạy cảm với p99.
- Bỏ qua nếu bạn cần JOIN, transaction đa partition, hay analytics ad-hoc.
- Cân nhắc kỹ vụ đổi giấy phép tháng 12/2024 và hệ sinh thái nhỏ trước khi cam kết.

Đang đánh giá nó cho một hệ thống mới? Hãy bắt đầu bằng việc viết ra danh sách truy vấn; lựa chọn database thường tự lộ ra từ chính bài tập đó. [INTERNAL-LINK: hướng dẫn query-first data modeling → bài tiếp theo trong series database]

<!-- NOTE TO AUTHOR: Bài không dùng marker [PERSONAL EXPERIENCE] vì chưa có kinh nghiệm production ScyllaDB nào được ghi nhận. Nếu bạn từng vận hành ScyllaDB, Cassandra hay MongoDB trong production (metrics, sự cố, chuyện tuning), gửi cho mình để bổ sung thành nội dung [PERSONAL EXPERIENCE] thật. -->

## Nguồn tham khảo

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
