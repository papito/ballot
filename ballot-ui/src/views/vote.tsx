import './vote.css'
import React, { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import Brand from '../components/brand.tsx'
import Footer from '../components/footer.tsx'
import GeneralError from '../components/general_error.tsx'

// https://github.com/axios/axios/discussions/5859
// eslint-disable-next-line import/named
import axios, { AxiosResponse } from 'axios'

function Vote(): React.JSX.Element {
    const [generalError, setGeneralError] = useState<string | null>(null)
    const params = useParams()
    // const possibleEstimates: Readonly<string[]> = ['?', '0', '1', '2', '3', '5', '8', '13', '20', '40', '100']

    const userId = params.userId
    console.assert(userId, 'userId is required')

    useEffect(() => {
        const fetchUser = async (): Promise<void> => {
            const response: AxiosResponse = await axios.get(`/api/user/${userId}`)
            console.log(response.data)
        }

        fetchUser().catch((error: unknown) => {
            setGeneralError(`An error occurred (${error}). See server logs.`)
        })
    }, [userId])

    return (
        <div id="Vote">
            <Brand />
            <GeneralError error={generalError} />
            <div id="voteContainer"></div>
            <Footer />
        </div>
    )
}

export default Vote
