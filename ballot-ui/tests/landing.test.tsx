import { render, screen } from '@testing-library/react'
import '@testing-library/jest-dom'
import { userEvent } from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { v4 as uuidv4 } from 'uuid'

// @ts-ignore
import Landing from '../src/views/landing.tsx'
// @ts-ignore
import { getUrlParams, mockServer } from './utils.ts'

mockServer.listen()

describe('Landing page tests', () => {
    beforeEach(() => {
        Object.defineProperty(window, 'assign', jest.fn)
    })

    it('loads the landing page without much drama', async () => {
        render(<Landing />)
        const newVotingSpaceBtnTxt = screen.getByPlaceholderText('Your name/alias')
        expect(newVotingSpaceBtnTxt).toBeInTheDocument()
    })

    it('creates a new voting space like it is nothing', async () => {
        const sessionId = uuidv4()
        const userId = uuidv4()
        console.log(sessionId)

        const handlers = [
            http.post('/api/session', () => {
                return HttpResponse.json({
                    id: sessionId,
                })
            }),

            http.post('/api/user', () => {
                return HttpResponse.json({
                    id: userId,
                })
            }),

            http.put('/api/vote/start', () => {
                return HttpResponse.json({})
            }),

            // http.post('/api/session', () => {
            //     return HttpResponse.json({ err: 'lol' }, { status: 400 })
            // }),
        ]

        mockServer.use(...handlers)

        render(<Landing />)
        const newVotingSpaceBtn = screen.getByRole('button')
        expect(newVotingSpaceBtn).toBeInTheDocument()

        await userEvent.click(newVotingSpaceBtn)

        // check that the redirect fired properly
        const urlParams: { [key: string]: string } = getUrlParams(window.location.href)
        expect(urlParams.session_id).toBe(sessionId)
        expect(urlParams.user_id).toBe(userId)
    })

    // it('vehemently objects to no name provided', async () => {
    //     server.listen()
    //
    //     render(<Landing />)
    //
    //     const newVotingSpaceBtn = screen.getByRole('button')
    //     expect(newVotingSpaceBtn).toBeInTheDocument()
    //
    //     fireEvent.click(newVotingSpaceBtn)
    // })
})
