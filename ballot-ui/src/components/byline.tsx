import './byline.css'
import React from 'react'

function Byline(): React.JSX.Element {
    return (
        <div id="byline">
            <div className="link">
                <div>
                    By{' '}
                    <a href="https://renegadeotter.com" target="_blank" rel="noreferrer">
                        Renegade Otter
                    </a>
                </div>
            </div>
            <div>
                <a href="https://renegadeotter.com" target="_blank" rel="noreferrer">
                    <img src="/renegade.png" alt="Renegade Otters" />
                </a>
            </div>
        </div>
    )
}

export default Byline
