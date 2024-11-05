import './copy_session_url.css'
import React from 'react'
import { CopyToClipboard } from 'react-copy-to-clipboard'
import { Session } from '../types/types.tsx'

interface CopyUrlProps {
    session: Session | null
}

function CopySessionUrl({ session }: CopyUrlProps): React.JSX.Element {
    const [clicked, setClicked] = React.useState(false)

    if (session === null) {
        return <></>
    }

    function timeout(delay: number): Promise<unknown> {
        return new Promise((res) => setTimeout(res, delay))
    }

    async function onCopyText(): Promise<void> {
        setClicked(true)
        await timeout(2000)
        setClicked(false)
    }

    return (
        <CopyToClipboard text={window.location.origin + '/vote/s/' + session.id} onCopy={onCopyText}>
            <div>
                <span className="text">
                    Use this to invite others here <i className="fas fa-arrow-alt-circle-right"></i>{' '}
                </span>
                <span>
                    <button className={'btn copy-url ' + (clicked ? 'clicked' : '')}>
                        <i className="fas fa-clipboard"></i>Session URL
                    </button>
                </span>
            </div>
        </CopyToClipboard>
    )
}

export default CopySessionUrl
