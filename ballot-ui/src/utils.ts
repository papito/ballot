import { Events } from './constants.ts'
import { TError } from './types/types.tsx'

export function isCustomEvent(event: Event): event is CustomEvent {
    return 'detail' in event
}

export function setGeneralError(msg: TError): void {
    const event = new CustomEvent(Events.generalError, { detail: { msg: msg } })
    document.dispatchEvent(event)
}

export function setFormError(msg: TError): void {
    const event = new CustomEvent(Events.formError, { detail: { msg: msg } })
    document.dispatchEvent(event)
}
