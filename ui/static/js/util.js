const domain = "http://localhost:4000";
const itemImgFSPath = "item/img/";

class Scanner {
  static code = "";
  static interval;
  static handleFunc;
  static title = document.querySelector("title");
  static titleTxt = document.querySelector("title").textContent;

  // process processes keydown characters
  static process(e) {
    if (Scanner.handleFunc === undefined) {
      console.error("Scanner.handleFunc is NOT initialized");
      return;
    }
    if (Scanner.interval) {
      clearInterval(Scanner.interval);
    }
    if (e.key == "Enter") {
      if (Scanner.code) {
        console.log(Scanner.code);
        Scanner.handleFunc(Scanner.code);
      }
      Scanner.code = "";
      return;
    }
    if (e.key != "Shift") {
      Scanner.code += e.key;
    }

    Scanner.interval = setInterval(function () {
      Scanner.code = "";
    }, 20);
  }

  // init initializes document's 'keydown' listenner on DOMContentLoaded, focus and remove on blur
  static init(func) {
    Scanner.handleFunc = func;

    if (Scanner.handleFunc === undefined) {
      console.error("Scanner.handleFunc is NOT initialized");
      return;
    }

    document.addEventListener("DOMContentLoaded", Scanner.start);

    window.onfocus = Scanner.start;
    window.onload = Scanner.start;
    window.focus();

    window.onblur = Scanner.stop;
  }

  static start() {
    Scanner.title.textContent = "QR - " + Scanner.titleTxt;
    console.log("QR on");

    Scanner.code = "";
    document.onkeydown = Scanner.process;
  }

  static stop() {
    Scanner.title.textContent = Scanner.titleTxt;
    console.log("QR off");

    Scanner.code = "";
    document.removeEventListener("keydown", Scanner.process);
  }
}

export { domain, Scanner };
