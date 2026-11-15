const http = require("http");

http.createServer((req, res) => {
    console.log("Backend 1 received:", req.url);
    res.end("Response from Backend 1");
}).listen(3001, () => console.log("Backend 1 running"));