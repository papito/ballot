import './brand.css'
import React from 'react'
import { Session } from '../features/vote.tsx'
import CopySessionUrl from './copy_session_url.tsx'

interface BrandProps {
    session: Session | null
}

function Brand({ session }: BrandProps): React.JSX.Element {
    return (
        <div id="Brand">
            <span>
                <a href="/">Ballot</a>
            </span>
            <span id="copySessionUrl">
                <CopySessionUrl session={session} />
            </span>
        </div>
    )
}

export default Brand
