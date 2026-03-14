// This server hosts the built React application as a lightweight product-style
// web service. It serves the compiled SPA directly and proxies platform API
// traffic to the Go control plane so the packaged stack behaves like a real
// deployment instead of a development-only Vite session.
import { createReadStream } from "node:fs";
import { access, stat } from "node:fs/promises";
import { createServer } from "node:http";
import { request as httpRequest } from "node:http";
import { request as httpsRequest } from "node:https";
import { extname, join, normalize } from "node:path";

const port = Number.parseInt(process.env.PORT ?? "3000", 10);
const host = process.env.HOST ?? "0.0.0.0";
const apiTarget = new URL(process.env.PLATFORM_API_PROXY_TARGET ?? "http://127.0.0.1:8080");
const distRoot = process.env.PLATFORM_WEB_DIST_ROOT ?? "/app/dist";

const mimeTypes = {
  ".css": "text/css; charset=utf-8",
  ".html": "text/html; charset=utf-8",
  ".ico": "image/x-icon",
  ".js": "application/javascript; charset=utf-8",
  ".json": "application/json; charset=utf-8",
  ".map": "application/json; charset=utf-8",
  ".png": "image/png",
  ".svg": "image/svg+xml; charset=utf-8",
  ".txt": "text/plain; charset=utf-8"
};

const server = createServer(async (req, res) => {
  try {
    if (!req.url) {
      res.writeHead(400, { "Content-Type": "text/plain; charset=utf-8" });
      res.end("missing request URL");
      return;
    }

    const requestURL = new URL(req.url, `http://${req.headers.host ?? "localhost"}`);
    if (requestURL.pathname === "/readyz") {
      res.writeHead(200, { "Content-Type": "application/json; charset=utf-8" });
      res.end(JSON.stringify({ status: "ok", service: "platform-web" }));
      return;
    }

    if (requestURL.pathname.startsWith("/api/") || requestURL.pathname === "/healthz") {
      proxyToAPI(req, res, requestURL);
      return;
    }

    await serveStaticAsset(requestURL.pathname, res);
  } catch (error) {
    res.writeHead(500, { "Content-Type": "text/plain; charset=utf-8" });
    res.end(error instanceof Error ? error.message : "unexpected platform-web error");
  }
});

server.listen(port, host, () => {
  console.log(`platform-web listening on http://${host}:${port}`);
});

function proxyToAPI(req, res, requestURL) {
  const transport = apiTarget.protocol === "https:" ? httpsRequest : httpRequest;
  const upstream = transport(
    {
      protocol: apiTarget.protocol,
      hostname: apiTarget.hostname,
      port: apiTarget.port,
      method: req.method,
      path: `${requestURL.pathname}${requestURL.search}`,
      headers: req.headers
    },
    (upstreamRes) => {
      res.writeHead(upstreamRes.statusCode ?? 502, sanitizeHeaders(upstreamRes.headers));
      upstreamRes.pipe(res);
    }
  );

  upstream.on("error", (error) => {
    res.writeHead(502, { "Content-Type": "application/json; charset=utf-8" });
    res.end(JSON.stringify({ error: "platform API proxy failure", detail: error.message }));
  });

  req.pipe(upstream);
}

async function serveStaticAsset(pathname, res) {
  const requestedPath = pathname === "/" ? "/index.html" : pathname;
  const normalizedPath = normalize(requestedPath).replace(/^(\.\.[/\\])+/, "");
  const resolvedPath = join(distRoot, normalizedPath);

  if (await fileExists(resolvedPath)) {
    await streamFile(resolvedPath, res);
    return;
  }

  await streamFile(join(distRoot, "index.html"), res);
}

async function streamFile(filePath, res) {
  const fileInfo = await stat(filePath);
  const extension = extname(filePath);

  res.writeHead(200, {
    "Content-Length": String(fileInfo.size),
    "Content-Type": mimeTypes[extension] ?? "application/octet-stream",
    "Cache-Control": extension === ".html" ? "no-cache" : "public, max-age=300"
  });

  createReadStream(filePath).pipe(res);
}

async function fileExists(path) {
  try {
    await access(path);
    return true;
  } catch {
    return false;
  }
}

function sanitizeHeaders(headers) {
  const output = {};
  for (const [key, value] of Object.entries(headers)) {
    if (value === undefined) {
      continue;
    }
    output[key] = value;
  }
  return output;
}
