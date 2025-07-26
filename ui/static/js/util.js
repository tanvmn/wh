const domain = "http://localhost:4000";

class Scanner {
  static code = "";
  static interval;
  static handleFunc;
  static title = document.querySelector("title");
  static titleTxt = document.querySelector("title").textContent;

  // qrListen process keydown characters
  static qrListen(e) {
    if (Scanner.handleFunc === undefined) {
      console.error("Scanner.handleFunc is NOT initialized");
      return;
    }
    if (Scanner.interval) {
      clearInterval(Scanner.interval);
    }
    if (e.key == "Enter") {
      if (Scanner.code) {
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

  // scan initializes document 'keydown' listenner on DOMContentLoaded, focus and remove on blur
  static scan() {
    if (Scanner.handleFunc === undefined) {
      console.error("Scanner.handleFunc is NOT initialized");
      return;
    }
    document.addEventListener("DOMContentLoaded", function () {
      Scanner.title.textContent = Scanner.titleTxt;
      Scanner.title.textContent = "QR - " + Scanner.titleTxt;
      document.onkeydown = Scanner.qrListen;
    });

    window.onfocus = function () {
      Scanner.title.textContent = Scanner.titleTxt;
      Scanner.title.textContent = "QR - " + Scanner.titleTxt;
      document.onkeydown = Scanner.qrListen;
    };

    window.onblur = function () {
      Scanner.title.textContent = Scanner.titleTxt;
      document.removeEventListener("keydown", Scanner.qrListen);
    };
  }
}

export { domain, Scanner };
