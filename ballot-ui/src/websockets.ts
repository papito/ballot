/**
 * This is using a specific Go/JS library for websockets:
 * https://github.com/desertbit/glue
 */

// eslint-disable-next-line @typescript-eslint/no-explicit-any
declare let glue: any

export default class Websockets {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    socket: any = null

    constructor() {
        this.socket = glue()
    }

    send(data: string): void {
        this.socket.send(data)
    }

    close(): void {
        this.socket.close()
    }

    reconnect(): void {
        this.socket.reconnect()
    }
}
