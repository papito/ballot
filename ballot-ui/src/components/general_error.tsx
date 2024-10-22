import './brand.css'
import React from 'react'
import './general_error.css'

export interface GeneralErrorProps {
    error: string | null
}

function GeneralError({ error }: GeneralErrorProps): React.JSX.Element {
    return error ? <div id="generalError" dangerouslySetInnerHTML={{ __html: error }}></div> : <div id="noError"></div>
}

export default GeneralError
