const http = require("http");

http.createServer((req, res) => {
    console.log("Backend 1 received:", req.url);

    setTimeout(() => {
        res.end("Response from Backend 1");
    }, 2000);

}).listen(3001, () => console.log("Backend 1 running"));