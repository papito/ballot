import '@testing-library/jest-dom'
// @ts-ignore
import { mockServer } from './utils.ts'
import 'jest-location-mock'

beforeAll(() => {
    mockServer.listen()
})

afterEach(() => {
    mockServer.resetHandlers()
})

afterAll(() => {
    mockServer.close()
})
