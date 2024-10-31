// eslint-disable-next-line @typescript-eslint/no-explicit-any
declare let glue: any

export default class Websockets {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    socket: any = null

    constructor() {
        this.socket = glue()

        // this.socket.on('connecting', function () {
        //     console.debug('Glue client connecting...')
        // })
        // this.socket.on('connected', () => {
        //     console.debug(`Glue client state: ${this.socket.state()}`)
        //     console.debug(`Glue client type: ${this.socket.type()}`)
        // })
    }

    send(data: string): void {
        this.socket.send(data)
    }
}
