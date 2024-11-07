import './byline.css'
import React from 'react'

function Byline(): React.JSX.Element {
    const url = 'https://renegadeotter.com'

    return (
        <div id="byline">
            <div className="link">
                <div>
                    By{' '}
                    <a href={url} target="_blank" rel="noreferrer">
                        Renegade Otter
                    </a>
                </div>
            </div>
            <div>
                <a href={url} target="_blank" rel="noreferrer">
                    <img src="/renegade.png" alt="Renegade Otters" />
                </a>
            </div>
        </div>
    )
}

export default Byline
