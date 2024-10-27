// eslint-disable-next-line @typescript-eslint/no-explicit-any
declare let glue: any

export default class Websockets {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    socket: any = null

    constructor() {
        this.socket = glue()
        console.log(`Glue client version: ${this.socket.version()}`)

        this.socket.on('connecting', function () {
            console.log('Glue client connecting...')
        })
        this.socket.on('connected', () => {
            console.log(`Glue client state: ${this.socket.state()}`)
            console.log(`Glue client type: ${this.socket.type()}`)
        })
    }

    send(data: string): void {
        this.socket.send(data)
    }
}
