'use client'
import {createContext, ReactNode, useContext, useReducer} from "react";

// Define the type for our context data
type DictationContextType = {
    state: DictationContextState,
    dispatch: React.Dispatch<Actions>,
};

// Define the type for our context data
type DictationContextState = {
    default: string,
    current: string
};

// Define type for each action type to enforce type safety
export type SetCurrentDictation = {
    type: ActionTypes.SET_CURRENT;
    payload: string;
};
// export type CompleteDictation = {
//     type: ActionTypes.COMPLETE;
// };

export type Actions =
    | SetCurrentDictation
    // | CompleteDictation

export const DictationContext = createContext<DictationContextType>({
    state: {
        default:'',
        current:'',
    },
    dispatch: ()=>null,
});

interface DictationContextProviderProps {
    children: ReactNode,
}

// Define action types as an enum to ensure consistency and prevent typos
export enum ActionTypes {
    SET_CURRENT = "SET_CURRENT",
    //COMPLETE = "COMPLETE",
}

// Reducer function
const reducer = (state: DictationContextState, action: Actions) => {
    switch (action.type) {
        case ActionTypes.SET_CURRENT:
            return {...state, current: action.payload};
        // case ActionTypes.COMPLETE:
        //     return {...state, current: state.default}
        default:
            return {...state, lastError: "unknown action type!"} // TODO: FIX!
    }
};

// TODO: USE THIS IN PLACES!!!! TRIAL IT ON PLATE!!!
export function DictationContextProvider({children}:DictationContextProviderProps){
    const [state, dispatch] = useReducer(reducer, {default:"main",current:"main"});

    return (
        <DictationContext.Provider value={{ state, dispatch }}>
            {children}
        </DictationContext.Provider>
    );
}

export function useDictationContext() {
    const context = useContext(DictationContext);

    if (!context) {
        throw new Error(
            'The DictationContext must be used within a Provider'
        );
    }

    return context;
}