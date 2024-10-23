import { render, screen } from '@testing-library/react'

import '@testing-library/jest-dom'
// @ts-ignore
import Landing from '../src/views/landing.tsx'

describe('Login component tests', () => {
    beforeEach(() => {})

    afterEach(() => {})

    it('Landing page renders correctly', async () => {
        render(<Landing />)
        const newVotingSpaceBtnTxt = screen.getByText('New Voting Space')
        expect(newVotingSpaceBtnTxt).toBeInTheDocument()
    })
})
