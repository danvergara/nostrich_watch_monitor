Nostr Monitor
=============

NIP-66 Relay Monitor Architecture

                          =====================================

┌─────────────────────────────────────────────────────────────────────────────────────┐
│                                   SCHEDULER LAYER                                   │
└─────────────────────────────────────────────────────────────────────────────────────┘

    ┌──────────────────────┐           ┌─────────────────────┐
    │   Go Job Scheduler   │           │     PostgreSQL      │
    │                      │◄──────────┤                     │
    │  • Cron-based timer  │   reads   │  • Relay list       │
    │  • Every 15 minutes  │   from    │  • Configuration    │
    │  • Creates job batch │           │  • Monitor state    │
    └──────────┬───────────┘           └─────────────────────┘
               │
               │ pushes jobs
               ▼

┌─────────────────────────────────────────────────────────────────────────────────────┐
│                                    QUEUE LAYER                                      │
└─────────────────────────────────────────────────────────────────────────────────────┘

                        ┌──────────────────────────┐
                        │      Redis Queue         │◄─── Job Status Updates
                        │                          │
                        │  Job Types:              │
                        │  ┌─────────────────────┐ │     ┌─────────────────────┐
                        │  │ relay_check_job     │ │     │   Nostr Network     │
                        │  │ {                   │ │     │                     │
                        │  │   url: "wss://..."  │ │     │ Published 30166     │
                        │  │   checks: [...],    │ │     │ events for online   │
                        │  │   timeouts: {...},  │ │◄────┤ relays              │
                        │  │   private_key: "..."│ │     │                     │
                        │  │   publish_relays:[] │ │     │ • relay.damus.io    │
                        │  │ }                   │ │     │ • nos.lol           │
                        │  └─────────────────────┘ │     │ • relay.nostr.band  │
                        │                          │     └─────────────────────┘
                        │  Features:               │
                        │  • Priority queues       │
                        │  • Job retry logic       │
                        │  • Dead letter queue     │
                        │  • Rate limiting         │
                        └──────────┬───────────────┘
                                   │
                                   │ pops jobs
                                   ▼

┌─────────────────────────────────────────────────────────────────────────────────────┐
│                                   WORKER LAYER                                      │
└─────────────────────────────────────────────────────────────────────────────────────┘

    ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
    │ Go Worker #1    │    │ Go Worker #2    │    │ Go Worker #N3   │
    │                 │    │                 │    │                 │
    │ • WebSocket     │    │ • WebSocket     │    │ • WebSocket     │
    │   health checks │    │   health checks │    │   health checks │
    │ • NIP-11 fetch  │    │ • NIP-11 fetch  │    │ • NIP-11 fetch  │
    │ • Signs 30166   │    │ • Signs 30166   │    │ • Signs 30166   │
    │   events        │    │   events        │    │   events        │
    │ • Publishes to  │    │ • Publishes to  │    │ • Publishes to  │
    │   Nostr relays  │    │   Nostr relays  │    │   Nostr relays  │
    │ • WaitGroup     │    │ • WaitGroup     │    │ • WaitGroup     │
    │ • Goroutines    │    │ • Goroutines    │    │ • Goroutines    │
    └─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
              │                      │                      │
              │ stores results       │ stores results       │ stores results
              │ & publishes 30166    │ & publishes 30166    │ & publishes 30166
              ▼                      ▼                      ▼

┌─────────────────────────────────────────────────────────────────────────────────────┐
│                                  STORAGE LAYER                                      │
└─────────────────────────────────────────────────────────────────────────────────────┘

    ┌─────────────────────┐                    ┌─────────────────────┐
    │    PostgreSQL       │                    │    Redis Cache      │
    │                     │                    │                     │
    │  Tables:            │                    │  • Recent results   │
    │  • relay_results    │                    │  • Job status       │
    │  • relays           │                    │  • Performance      │
    │  • monitor_config   │                    │    metrics          │
    └─────────┬───────────┘                    └─────────┬───────────┘
              │                                          │
              │ feeds data to                            │ feeds data to
              ▼                                          ▼

┌─────────────────────────────────────────────────────────────────────────────────────┐
│                                 PUBLISH LAYER                                       │
└─────────────────────────────────────────────────────────────────────────────────────┘

    ┌──────────────────────┐                    ┌─────────────────────┐
    │   Go Coordinator     │                    │    Web Dashboard    │
    │                      │                    │                     │
    │  • Publishes 10166   │                    │  • Website UI       │
    │    announcements     │                    │  • API endpoints    │
    │  • Cleanup tasks     │                    │  • Relay search     │
    │  • System monitoring │                    │  • Historical data  │
    │  • Database          │                    │  • Performance      │
    │    maintenance       │                    │    charts           │
    └──────────────────────┘                    └─────────────────────┘


                                  DATA FLOW
                                 ═══════════

    ┌─────────┐    ┌─────────┐    ┌─────────────────────┐    ┌─────────┐
    │   15    │───▶│  Queue  │───▶│ Workers Check &     │───▶│ Results │
    │ minute  │    │  Jobs   │    │ Publish 30166       │    │  Store  │
    │  tick   │    │         │    │ Events Immediately  │    │         │
    └─────────┘    └─────────┘    └─────────────────────┘    └─────────┘


                                JOB FLOW DETAIL
                               ═════════════════

Timer ──┐
        │
        ▼
    ┌───────────────────────────────────────────────────────────┐
    │              Go Scheduler Process                         │
    │                                                           │
    │  1. Load enabled relays from PostgreSQL                   │
    │  2. Create job for each relay                             │
    │  3. Push jobs to Redis queue                              │
    │  4. Monitor job completion                                │
    │  5. Trigger publisher when cycle completes                │
    └─────────────────────┬─────────────────────────────────────┘
                          │
                          ▼
                    Redis Queue
                          │
                          ▼
    ┌───────────────────────────────────────────────────────────┐
    │                Go Worker Pool                             │
    │                                                           │
    │  Worker Process:                                          │
    │  1. Pop job from Redis                                    │
    │  2. Parse job data (relay URL, timeouts, checks)          │
    │  3. Run health checks concurrently:                       │
    │     • WebSocket open test                                 │
    │     • Read capability test                                │
    │     • Write capability test                               │
    │     • NIP-11 document fetch                               │
    │  4. IF RELAY IS ONLINE:                                   │
    │     • Create & sign 30166 event                           │
    │     • Publish to target Nostr relays                      │
    │  5. Store results in PostgreSQL                           │
    │  6. Update Redis with job completion                      │
    └───────────────────────────────────────────────────────────┘


                              SCALING POINTS
                             ═══════════════

    Horizontal Scaling:
    • Add more Go worker instances
    • Redis handles load balancing
    • Each worker is stateless

    Vertical Scaling:
    • Increase worker concurrency per instance
    • Tune Redis memory/persistence
    • Scale PostgreSQL connections

    Geographic Scaling:
    • Deploy workers in different regions
    • Use region-specific job queues
    • Aggregate results centrally


                              TECHNOLOGY STACK
                             ═══════════════

    Go Components:
    • github.com/robfig/cron/v3 (scheduler)
    • github.com/go-redis/redis/v8 (queue client)
    • github.com/lib/pq (PostgreSQL driver)
    • github.com/nbd-wtf/go-nostr (Nostr publishing)

    Infrastructure:
    • Redis 8.x (job queue + cache)
    • PostgreSQL 15+ (persistent storage)
    • Optional: Docker containers
    • Optional: Kubernetes orchestration
