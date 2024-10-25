import { setupServer } from 'msw/node'

export const mockServer = setupServer()

export function getUrlParams(url: string): { [key: string]: string } {
    const params: { [key: string]: string } = {}
    const queryString = url.split('?')[1]
    if (!queryString) {
        return params
    }

    const pairs = queryString.split('&')
    for (const pair of pairs) {
        const [key, value] = pair.split('=')
        if (key) {
            params[decodeURIComponent(key)] = decodeURIComponent(value || '')
        }
    }

    return params
}
