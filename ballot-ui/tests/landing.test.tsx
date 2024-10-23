import { render, screen, fireEvent } from '@testing-library/react'
import { v4 as uuidv4 } from 'uuid'

import '@testing-library/jest-dom'
import { http, HttpResponse } from 'msw'
import { setupServer } from 'msw/node'
// @ts-ignore
import Landing from '../src/views/landing.tsx'

export const handlers = [
    http.post('/api/session', () => {
        return HttpResponse.json({
            id: uuidv4(),
        })
    }),
    http.post('/api/user', () => {
        return HttpResponse.json({
            id: uuidv4(),
        })
    }),
    http.put('/api/vote/start', () => {
        return HttpResponse.json({})
    }),
]

const server = setupServer(...handlers)

server.listen()

describe('Landing page tests', () => {
    it('loads the landing page without much drama', async () => {
        render(<Landing />)
        const newVotingSpaceBtnTxt = screen.getByPlaceholderText('Your name/alias')
        expect(newVotingSpaceBtnTxt).toBeInTheDocument()
    })

    it('creates a new voting space like it is nothing', async () => {
        render(<Landing />)

        const newVotingSpaceBtn = screen.getByRole('button')
        expect(newVotingSpaceBtn).toBeInTheDocument()

        fireEvent.click(newVotingSpaceBtn)
    })
})
