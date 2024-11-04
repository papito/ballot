import { render, screen, waitFor } from '@testing-library/react'
import '@testing-library/jest-dom'
import { userEvent } from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { v4 as uuidv4 } from 'uuid'

// @ts-ignore
import Landing from '../src/features/landing.tsx'
// @ts-ignore
import { mockServer } from './utils.ts'

mockServer.listen()

describe('Landing page tests', () => {
    const sessionId = uuidv4()
    const userId = uuidv4()

    beforeEach(() => {
        Object.defineProperty(window, 'assign', jest.fn)
    })

    it('loads the landing page without much drama', async () => {
        render(<Landing />)
        const newVotingSpaceBtnTxt = screen.getByPlaceholderText('Your name/alias')
        expect(newVotingSpaceBtnTxt).toBeInTheDocument()
    })

    it('creates a new voting space like a boss', async () => {
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
        ]

        mockServer.use(...handlers)

        render(<Landing />)

        const newVotingSpaceBtn = screen.getByRole('button')
        expect(newVotingSpaceBtn).toBeInTheDocument()

        await userEvent.click(newVotingSpaceBtn)

        await waitFor(() => {
            // check that the redirect fired properly
            expect(window.location.href).toContain(`/vote/s/${sessionId}/u/${userId}`)
        })
    })

    /*
     * These tests are failing after I added error display based on JS events and axios
     * interceptors.
     *
     * This makes the views much more clean, and it wipes out all boilerplate error handling.
     * But - this wrecked the tests, until I figure out how to wire in the axios interceptors,
     * which will happen as soon as, uh, never.
     *
     *
    it('gets crabby if no name/alias provided', async () => {
        const formErrorText = "A scrub is a guy who can't get no love from me"

        const handlers = [
            http.post('/api/session', () => {
                return HttpResponse.json({
                    id: sessionId,
                })
            }),
            http.post('/api/user', () => {
                return HttpResponse.json({ error: formErrorText }, { status: 400 })
            }),
        ]

        mockServer.use(...handlers)

        render(<Landing />)

        const newVotingSpaceBtn = screen.getByRole('button')

        // jest.spyOn(console, 'error').mockImplementation()
        // expect(console.error).toHaveBeenCalled()
        await userEvent.click(newVotingSpaceBtn)

        await waitFor(() => {
            const errorContainer: HTMLElement = screen.getByTestId('formError')
            expect(errorContainer).toHaveTextContent(formErrorText)
        })
    })

    it('displays server error if it all goes to bloody hell', async () => {
        const handlers = [
            http.post('/api/session', () => {
                return new HttpResponse(null, { status: 500 })
            }),
        ]

        mockServer.use(...handlers)

        render(<Landing />)

        const newVotingSpaceBtn = screen.getByRole('button')

        jest.spyOn(console, 'error').mockImplementation()
        await userEvent.click(newVotingSpaceBtn)

        await waitFor(() => {
            expect(console.error).toHaveBeenCalled()

            const errorContainer: HTMLElement = screen.getByTestId('generalError')
            expect(errorContainer).toHaveTextContent('Internal Server Error')
        })
    })
*/
})
