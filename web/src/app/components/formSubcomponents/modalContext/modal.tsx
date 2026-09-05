'use client'

import {createContext, ReactNode, useContext, useEffect, useReducer, useState} from "react";
import * as React from "react";
// Define the type for our context data
type ModalContextType = {
    state: PopupModalInfo,
    dispatch: React.Dispatch<Actions>,
};

export interface PopupInfo {
    header: string
    text?: string
    isErr: boolean // TODO: use this!
}
export const DefaultPopupInfo: PopupInfo = {
    header: "initial header",
    text: "you should never see this text",
    isErr: false,
}

export const ModalContext = createContext<ModalContextType>({
    state: {info: DefaultPopupInfo, open: false},
    dispatch: () => null,
})
export function useModalContext() {
    const context = useContext(ModalContext);

    if (!context) {
        throw new Error(
            'The ModalContext must be used within an ModalContextProvider'
        );
    }

    return context;
}
interface ModalContextProviderProps {
    children: ReactNode,
}
interface PopupModalInfo {
    info: PopupInfo
    open: boolean
}
// Reducer function
const reducer:(state: PopupModalInfo, action: Actions)=>PopupModalInfo = (state: PopupModalInfo, action: Actions) => {
    switch (action.type) {
        case ActionTypes.SET_MODAL_INFO:
            if(!action.payload){
                return {info: action.payload||{
                        header: "Error in reducer",
                        text: "undefined payload",
                        isErr: true,
                    }, open: true}
            } else {
                return {info: action.payload, open: true}
            }
        case ActionTypes.CLOSE_MODAL:
            return {...state, open: false}
        default:
            return {info: {
                    header: "Error in reducer",
                    text: "bad action type",
                    isErr: true,
                }, open: true}
    }
};
export const ModalContextProvider = ({children}:ModalContextProviderProps) => {
    const [state, dispatch] = useReducer(reducer, {info:DefaultPopupInfo,open:false});

    return (
        <ModalContext.Provider value={{ state, dispatch }}>
            <PopupApp/>
            {children}
        </ModalContext.Provider>
    );
}
// export function ModalProvider(props:React.PropsWithChildren<{}>){
//     const ctx = useContext(ModalContext) // TODO: ensure ok
//     return <ModalContext.Provider value={ctx}>
//         {props.children}
//     </ModalContext.Provider>
// }
export enum ActionTypes {
    CLOSE_MODAL = "CLOSE_MODAL",
    SET_MODAL_INFO = "SET_MODAL_INFO"
}
export type CloseModalAction = {
    type: ActionTypes.CLOSE_MODAL;
};
export type SetModalInfoAction = {
    type: ActionTypes.SET_MODAL_INFO;
    payload?: PopupInfo;
};
export type Actions =
    | CloseModalAction
    | SetModalInfoAction

export function PopupApp() {
    const {state,dispatch} = useModalContext();
    const close = (e:React.MouseEvent<HTMLButtonElement, MouseEvent>)=>{
        e.stopPropagation()
        e.preventDefault()
        dispatch({ type: ActionTypes.CLOSE_MODAL });
    }
    if (state.open) {
        return <div className={"popupModal"}>
            <div className={"popupModalContent"}>
                <h3>{state.info.header}</h3>
                <p className={state.info.isErr?"error":""}>{state.info.text || "no text set, you should never see this message"}</p>
                <button className={"basicButton buttonFullWidth"} onClick={close}>{"Close"}</button>
            </div>
        </div>
    }
    return null
}