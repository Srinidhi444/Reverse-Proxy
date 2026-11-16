const http = require("http");

http.createServer((req, res) => {
    console.log("Backend 2 received:", req.url);

    setTimeout(() => {
        res.end("Response from Backend 2");
    }, 2000);

}).listen(3002, () => console.log("Backend 2 running"));