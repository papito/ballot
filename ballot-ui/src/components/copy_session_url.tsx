import './copy_session_url.css'
import React from 'react'
import { CopyToClipboard } from 'react-copy-to-clipboard'
import { Session } from '../views/vote.tsx'

interface CopyUrlProps {
    session: Session
}

function CopySessionUrl({ session }: CopyUrlProps): React.JSX.Element {
    const [clicked, setClicked] = React.useState(false)

    function timeout(delay: number): Promise<unknown> {
        return new Promise((res) => setTimeout(res, delay))
    }
    async function onCopyText(): Promise<void> {
        setClicked(true)
        await timeout(2000)
        setClicked(false)
    }

    return (
        <div id="copySessionUrl">
            <CopyToClipboard text={window.location.origin + '/p/vote/s/' + session.id} onCopy={onCopyText}>
                <button className={'btn copy-url ' + (clicked ? 'clicked' : '')}>
                    <i className="fas fa-clipboard"></i>Copy session URL
                </button>
            </CopyToClipboard>
        </div>
    )
}

export default CopySessionUrl
