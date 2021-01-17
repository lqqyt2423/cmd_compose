const pid = process.pid;

let count = 1;

setInterval(() => {
    console.log(`${pid} ${new Date()} ${count++} times`);
}, 1000);
