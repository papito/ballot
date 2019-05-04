declare var glue: any;

export default class Websockets {
  socket: any = null;

  constructor() {
    this.socket = glue();
    console.log(`Glue client version: ${this.socket.version()}`);

    this.socket.on("connecting", function () {
      console.log("Glue client connecting...")
    });
    this.socket.on("connected", () => {
      console.log(`Glue client state: ${this.socket.state()}`);
      console.log(`Glue client type: ${this.socket.type()}`);
    });
  }

  send(data: string) {
    this.socket.send(data);
  }
};