'use client'

import {createContext, ReactNode, useContext, useReducer} from 'react';
import { useCookieConsent } from "react-cookie-manager";
import { useCookies } from 'react-cookie';

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
    // const {detailedConsent}= useCookieConsent() // TODO: remove if not works
    // const [cookies, setCookie] = useCookies(['mostRecentRfidReader']);
    switch (action.type) {
        case ActionTypes.SET_READER:
            // SeT VALUES IN STORAGE OR COOKIES IF AVAILABLE
            // if (detailedConsent!==null && state.selected !== action.payload && action.payload !== undefined) {
            //     if (detailedConsent.FunctionalCookies.consented) {
            //         setCookie('mostRecentRfidReader', action.payload, {
            //
            //             path: '/',               // Accessible across your entire site
            //             maxAge: 60*60*24*7,          // Cookie expires in 7 days (in seconds)
            //             secure: true,            // Transmitted only over HTTPS
            //             sameSite: 'lax'          // Protection against CSRF attacks // TODO: fix? was lax, could need to be strict?
            //             // TODO: DO THIS? domain?: string;
            //             // TODO: DO THIS? httpOnly?: boolean;
            //             // TODO: DO THIS? partitioned?: boolean;
            //         })
            //     }
            //     if (typeof window !== 'undefined') {
            //         // Local storage
            //         if (detailedConsent.FunctionalLocalStorage.consented) {
            //             localStorage.setItem('mostRecentRfidReader', action.payload)
            //         }
            //         // Session Storage
            //         if (detailedConsent.FunctionalSessionStorage.consented) {
            //             sessionStorage.setItem('mostRecentRfidReader', action.payload)
            //         }
            //     }
            // }
            return {...state, selected: action.payload};
        case ActionTypes.SET_LAST_READ_TAG:
            return {...state, lastReadTag: action.payload}
        case ActionTypes.SET_ERROR:
            return {...state, lastError: action.payload}
        case ActionTypes.SET_LAST_READER:
            return {...state, lastReaderUsed: action.payload} // TODO: UNUSED
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
    //const [cookies] = useCookies(['mostRecentRfidReader']);
    // TODO: consider using useState in here instead of the reducer stuff????
    //const {detailedConsent}= useCookieConsent()
    if (!context) {
        throw new Error(
            'The ReaderOptionsContext must be used within an ReaderOptionsContextProvider'
        );
    }
    // try to set the most recently used reader on spin-up?
    // const getUserMostRecentlyUsedReader = new Promise<string>((accept,reject)=>{ // TODO:???
    //     // if (detailedConsent === null || detailedConsent === undefined) {
    //     //     reject("No consent for functional storage provided")
    //     //     return
    //     // } else {
    //     //     // if (detailedConsent.FunctionalCookies.consented) {
    //     //     //     const c:string |undefined|null= cookies.mostRecentRfidReader
    //     //     //     if (!(c===null||c===undefined||c==="")){
    //     //     //         accept(c)
    //     //     //         return
    //     //     //     }
    //     //     // }
    //     //     // if (detailedConsent.FunctionalLocalStorage.consented) {
    //     //     //     // Check Local storage
    //     //     //     if (typeof window !== 'undefined') {
    //     //     //         let val = localStorage.getItem('mostRecentRfidReader')
    //     //     //         if (val!==null){
    //     //     //             accept(val)
    //     //     //             return
    //     //     //         }
    //     //     //     }
    //     //     // }
    //     //     // if (detailedConsent.FunctionalSessionStorage.consented) {
    //     //     //     // Check Session Storage
    //     //     //     if (typeof window !== 'undefined') {
    //     //     //         let val = sessionStorage.getItem('mostRecentRfidReader')
    //     //     //         if (val!==null){
    //     //     //             accept(val)
    //     //     //             return
    //     //     //         }
    //     //     //     }
    //     //     // }
    //         reject("nothing found in storage or not allowed")
    //         return
    //     //}
    // })
    // getUserMostRecentlyUsedReader.then((reader)=>{ // TODO: validate works
    //     // TODO: ensure value is in options
    //     if (context.state.options.includes(reader)){
    //         context.dispatch({
    //             type: ActionTypes.SET_READER,
    //             payload: reader,
    //         });
    //     } else {
    //         throw "users historical reader ("+reader+") not an option"
    //     }
    // }).catch((e)=>{
    //     if (JSON.stringify(e) == "nothing found in storage or not allowed"){
    //         return context
    //     }
    //     console.error(JSON.stringify(e));
    // })
    return context;
}