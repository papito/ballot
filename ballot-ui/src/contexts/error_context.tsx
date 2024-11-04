/*
    This file is used to manage errors in the application.
    It uses axios interceptors to catch errors from the server and display them to the user.
    It also listens for custom events dispatched from other parts of the application and displays them to the user.
 */

// https://github.com/axios/axios/discussions/5859
// eslint-disable-next-line import/named
import axios, { AxiosError, AxiosResponse } from 'axios'
import React, { createContext, Dispatch, ReactNode, SetStateAction, useContext, useMemo, useState } from 'react'
import { TError, TValidationError } from '../types/types.tsx'

interface ErrorContextType {
    generalError: TError
    setGeneralError: Dispatch<SetStateAction<TError>>
    formError: TError
    setFormError: Dispatch<SetStateAction<TError>>
}

const ErrorContext = createContext<ErrorContextType>({
    generalError: null,
    setGeneralError: () => {},
    formError: null,
    setFormError: () => {},
})

const ErrorContextProvider = ({ children }: { children: ReactNode }): React.JSX.Element => {
    const [generalError, setGeneralError] = useState<TError>(null)
    const [formError, setFormError] = useState<TError>(null)

    axios.interceptors.request.use(
        (config) => {
            // reset all messages on new request
            setFormError(null)
            setGeneralError(null)
            return config
        },
        (error) => {
            setGeneralError('Error: Request failed')
            return Promise.reject(error)
        }
    )

    const onResponse = (response: AxiosResponse): AxiosResponse => {
        return response
    }

    const onResponseError = (axiosError: AxiosError<TValidationError>): Promise<AxiosError> => {
        switch (axiosError.response?.status) {
            case 400:
                setFormError(axiosError.response?.data.error)
                break
            default:
                setGeneralError(axiosError.response?.statusText || 'Error: HTTP' + axiosError.response?.status)
        }
        return Promise.reject(axiosError)
    }

    axios.interceptors.response.use(onResponse, onResponseError)

    const value = useMemo(
        () => ({ generalError, setGeneralError, formError, setFormError }),
        [generalError, setGeneralError, formError, setFormError]
    )

    return <ErrorContext.Provider value={value}>{children}</ErrorContext.Provider>
}

const useErrorContext = (): ErrorContextType => {
    const context = useContext(ErrorContext)

    if (!context) {
        throw new Error('useErrorContext must be used within an ErrorProvider')
    }
    return context
}

export { ErrorContext, ErrorContextProvider, useErrorContext }
