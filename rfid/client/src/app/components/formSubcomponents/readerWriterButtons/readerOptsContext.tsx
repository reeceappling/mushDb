'use client'

import {createContext, ReactNode, useContext, useReducer, useState} from 'react';
import {ReadRfidTag} from "@/app/view/[itemType]/[idEncoded]/serverActions";

export type ModalInfo = {
    modalType: string
    recordId: string
};

interface readerSelectorContext {
    options: string[]
    selected?: string
    lastReadTag?: string
    lastReaderUsed?: string
    lastError?: string
    // modalInfo?: ModalInfo
    // modalHistory?: ModalInfo  // TODO: MODAL HISTORY? MAKE NON-OPTIONAL!
}

// Define the type for our context data
type RfidReaderContextType = {
    state: readerSelectorContext,
    dispatch: React.Dispatch<Actions>,
};

export const ReaderOptionsContext = createContext<RfidReaderContextType>({
    state: {
        options: ["initA"], // TODO: ?
    },
    dispatch: ()=>null,
});

// Define action types as an enum to ensure consistency and prevent typos
export enum ActionTypes {
    CLEAR_ERROR = "CLEAR_ERROR",
    SET_READER = "SET_READER",
    SET_LAST_READER = "SET_LAST_READER",
    SET_ERROR = "ERROR",
    SET_LAST_READ_TAG = "SET_LAST_READ_TAG",
    SET_MODAL_INFO = "SET_MODAL_INFO"
}

// Define type for each action type to enforce type safety
export type SetReaderAction = {
    type: ActionTypes.SET_READER;
    payload?: string;
};
export type SetLastReadTagAction = {
    type: ActionTypes.SET_LAST_READ_TAG;
    payload?: string;
};
export type SetLastReaderAction = {
    type: ActionTypes.SET_LAST_READER;
    payload?: string;
};
export type SetErrorAction = {
    type: ActionTypes.SET_ERROR;
    payload?: string;
};
export type ClearErrorAction = {
    type: ActionTypes.CLEAR_ERROR;
    payload?: string;
};
export type SetModalInfoAction = {
    type: ActionTypes.SET_MODAL_INFO;
    payload?: ModalInfo;
};

// Define a union type Actions to represent all possible action types
export type Actions =
    | ClearErrorAction
    | SetReaderAction
    | SetLastReadTagAction
    | SetLastReaderAction
    | SetModalInfoAction
    | SetErrorAction;

// Reducer function
const reducer = (state: readerSelectorContext, action: Actions) => {
    switch (action.type) {
        case ActionTypes.SET_READER:
            return {...state, selected: action.payload};
        case ActionTypes.SET_LAST_READ_TAG:
            return {...state, lastReadTag: action.payload}
        // case ActionTypes.SET_MODAL_INFO:
        //     return {...state, modalInfo: action.payload}
        case ActionTypes.SET_ERROR:
            return {...state, lastError: action.payload}
        case ActionTypes.SET_LAST_READER:
            return {...state, lastReaderUsed: action.payload}
        case ActionTypes.CLEAR_ERROR:
            return {...state, lastError: undefined}
        default:
            return {...state, lastError: "unknown action type!"}
    }
};

interface ReaderOptionsContextProviderProps {
    children: ReactNode,
    initialState: readerSelectorContext,
}

export const ReaderOptionsContextProvider = ({children, initialState}:ReaderOptionsContextProviderProps) => {
    const [state, dispatch] = useReducer(reducer, initialState);

    return (
        <ReaderOptionsContext.Provider value={{ state, dispatch }}>
            {children}
        </ReaderOptionsContext.Provider>
    );
}

export function useRfidReaderContext() {
    const context = useContext(ReaderOptionsContext);

    if (!context) {
        throw new Error(
            'The App Context must be used within an AppContextProvider'
        );
    }

    return context;
}