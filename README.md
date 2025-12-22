# Hackathons API

A powerful, unified API that aggregates **India-specific hackathons** from 9+ major platforms in real-time. Built with Go, it provides a clean, deduplicated, and sorted list of upcoming opportunities for developers.

## 🚀 Features

-   **Multi-Platform Aggregation**: Fetches hackathons from:
    -   Devfolio
    -   Unstop
    -   Devpost
    -   MLH (Major League Hacking)
    -   HackerEarth
    -   Hack2Skill
    -   ReSkill
    -   WhereUElevate
    -   Devnovate
-   **Smart Filtering**:
    -   Auto-filters non-India events (unless global/online).
    -   Removes expired or past events.
    -   Deduplicates events listed on multiple platforms.
-   **Robust Data**:
    -   Dynamic date parsing (handles various formats and timezones).
    -   Auto-detects upcoming seasons (e.g., MLH 2026).
    -   Graceful handling of missing data (no dummy dates).
-   **High Performance**:
    -   Concurrent fetching with Goroutines.
    -   In-memory caching (expires every 1 hour).
    -   Resilient headers and retries for scraping.

## 🛠️ Tech Stack

-   **Language**: Go (Golang) 1.23+
-   **Router**: Chi
-   **Scraping**: Colly, GoQuery
-   **Architecture**: Modular (Fetchers -> Aggregator -> API Handler)

## 📦 Installation & Run

1.  **Clone the repository**
    ```bash
    git clone https://github.com/0xarchit/hackathon-api.git
    cd hackathon-api
    ```

2.  **Install Dependencies**
    ```bash
    go mod tidy
    ```

3.  **Run the Server**
    ```bash
    go run cmd/server/main.go
    ```
    The server will start on `http://localhost:8080`.

## 📡 API Endpoints

### 1. Get Hackathons
Fetch the list of aggregated hackathons.

-   **Endpoint**: `GET /api/v1/hackathons`
-   **Query Parameters**:
    -   `platform` (optional): Filter by platform (e.g., `devfolio`, `mlh`).
    -   `mode` (optional): Filter by mode (`online` or `offline`).
-   **Example**:
    ```http
    GET http://localhost:8080/api/v1/hackathons?mode=online
    ```
-   **Response**:
    ```json
    {
        "status": "success",
        "count": 15,
        "generated_at": "2025-12-23T10:00:00Z",
        "data": [
            {
                "id": "devfolio-123",
                "title": "HackThis 2025",
                "platform": "devfolio",
                "mode": "online",
                "location": "Online",
                "country": "India",
                "start_date": "2025-12-25T09:00:00Z",
                "end_date": "2025-12-27T09:00:00Z",
                "registration_end": "2025-12-24T23:59:00Z",
                "url": "https://hackthis.devfolio.co"
            }
        ]
    }
    ```

### 2. Health Check
Check if the API is running.

-   **Endpoint**: `GET /health`
-   **Response**:
    ```json
    {
        "status": "healthy",
        "time": "2025-12-22T20:30:00+05:30"
    }
    ```

## ☁️ Deployment

### Option 1: Vercel (Serverless)
**Note**: Vercel has a 10s timeout on the free tier. Since this API scrapes 9 sites in real-time on the first load, it might time out on "Cold Starts".
1.  Install Vercel CLI: `npm i -g vercel`
2.  Run `vercel` in the project root.
3.  Deploy!

### Option 2: Render / Railway (Docker) - **Recommended**
Since this API benefits from in-memory caching to avoid repeatedly scraping sites, a persistent instance is better.
1.  Push code to GitHub.
2.  Connect repository to **[Render](https://render.com)** or **[Railway](https://railway.app)**.
3.  Select "Docker" as the build type.
4.  It will automatically use the `Dockerfile` and deploy.

## 🤝 Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
