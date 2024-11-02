import './footer.css'
import React from 'react'

function Footer(): React.JSX.Element {
    return (
        <div className="footer">
            <div className="renegade"></div>
            <div className="sauce">
                <a href="https://github.com/papito/ballot" target="_blank" rel="noreferrer">
                    <img src="/gh.png" alt="GitHub"/>
                </a>
            </div>
        </div>
    )
}

export default Footer
