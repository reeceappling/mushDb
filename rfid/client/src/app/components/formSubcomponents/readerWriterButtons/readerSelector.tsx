import * as React from "react";
import {SyntheticEvent} from "react";
import {
    Actions,
    ActionTypes,
    useRfidReaderContext
} from "@/app/components/formSubcomponents/readerWriterButtons/readerOptsContext";
import {Makeid} from "@/app/components/TopBar";


interface rfidSelectorProps {
    defaultOption?: string,
    txt?: string,
    onSelect?: (s?: string) => void,
    headerLevel?: number
}

export default function ReaderWriterSelector(props:rfidSelectorProps) {
    const {state, dispatch} = useRfidReaderContext()
    const defaultOpt = props.defaultOption || "none"
    const onSelect = (e: SyntheticEvent<HTMLSelectElement, Event>) => {
        let val = e.currentTarget.value
        if (val && val !== state.selected) {
            dispatch({
                type: ActionTypes.SET_READER,
                payload: (val==defaultOpt)?undefined:val
            })
        }
        props.onSelect && props.onSelect(val)
    }
    return <div className={"centerH gapTop rwSelector"}>{props.txt || "Current Reader/Writer"}<select className={"tailwindSelector"} value={state.selected || defaultOpt} onChange={onSelect}>
        {[defaultOpt, ...state.options].map(function (name, i) {
            return <option value={name} key={i}>{name}</option>
        })}
    </select></div>
}

export function ReadTagFunc(dispatch: React.Dispatch<Actions>, session?: string, selectedReader?: string): Promise<string> {
    return new Promise((resolve, reject) => {
        // TODO: fix sess
        // ReadRfidTag(session, dispatch, selectedReader).then((id)=>{
        //     resolve(id)
        // },(err)=>{
        //     reject(err)
        // })
        // TODO: remove all below when reenabled
        if (!selectedReader) {
            let toWrite = "no rfid reader selected"
            dispatch({
                type: ActionTypes.SET_ERROR,
                payload: toWrite,
            })
            reject(toWrite)
            return
        }
        let readerName = selectedReader
        let tagVal = Makeid(5)
        if (readerName === "goodTestRfid"){
            tagVal = "4Wj8HxCMmcs" // TODO: Test empty plate id
        }
        dispatch({
            type: ActionTypes.SET_LAST_READ_TAG,
            payload: tagVal,
        })
        dispatch({
            type: ActionTypes.SET_LAST_READER,
            payload: readerName,
        })
        dispatch({
            type: ActionTypes.CLEAR_ERROR,
        })
        resolve(tagVal)
    })
}

export type selectReaderResult = {
    didRead: boolean;
    payload?: string;
};

export function SelectReaderFunc(dispatch: React.Dispatch<Actions>, doRead: boolean, session?: string, reader?: string): Promise<selectReaderResult> {
    let out:selectReaderResult = {didRead: doRead}
    return new Promise<selectReaderResult>((resolve, reject) => {
        dispatch({
            type: ActionTypes.SET_READER,
            payload: reader,
        })
        if (!doRead) {
            ReadTagFunc(dispatch, session, reader).then(id => {
                resolve({didRead: true, payload: id})
                return
            }, err => {
                reject(err)
                return
            })
        } else {
            resolve({didRead: false})
        }
    })
}

export function ReadRFIDButton(
    {
        handleTagRead, txt, session
    }:{
        txt?:string,
        handleTagRead:(id: string)=>void
        session?: string
    }) {
    const {state, dispatch} = useRfidReaderContext()
    return <button className={"basicButtonSmall"} onClick={()=>{
        ReadTagFunc(dispatch, session, state.selected).then(handleTagRead)
    }}>{txt || "Read ID from RFID Reader"}</button>
}

export function RfidSelectorWithReadButton( // TODO: use????
    {
        defaultReaderOption, readerWriterTxt, onWriterSelect, readButtonTxt, handleTagRead, headerLevel,autoRead
    }:{
        defaultReaderOption?: string,
        readerWriterTxt?: string,
        onWriterSelect?: (s?: string) => void,
        headerLevel?: number,
        readButtonTxt?:string,
        handleTagRead:(id: string)=>void,
        autoRead?:boolean,
    }){
    return <div>
        <ReaderWriterSelector txt={readerWriterTxt} headerLevel={headerLevel} defaultOption={defaultReaderOption} onSelect={onWriterSelect} />
        <ReadRFIDButton handleTagRead={handleTagRead} txt={readButtonTxt} />
    </div>
}

// export function RfidSelectorSplitFromStateWithAutoread( // TODO: use????
//     {
//         defaultReaderOption, readerWriterTxt, onWriterSelect, readButtonTxt, handleTagRead, headerLevel
//     }:{
//         defaultReaderOption?: string,
//         readerWriterTxt?: string,
//         onWriterSelect?: (s?: string) => void,
//         headerLevel?: number,
//         readButtonTxt?:string,
//         handleTagRead:(id: string)=>void,
//     }){
//     const {state, dispatch} = useRfidReaderContext()
//     const defaultOpt = defaultReaderOption || "none"
//     // TODO: may need to move due to state
//     const [currentOption, setCurrentOption] = useState(defaultOpt)
//
//     const onSelect = (e: SyntheticEvent<HTMLSelectElement, Event>) => {
//         let val = e.currentTarget.value
//         if (val && val !== state.selected) {
//             dispatch({
//                 type: ActionTypes.SET_READER,
//                 payload: (val==defaultOpt)?undefined:val
//             })
//         }
//         setCurrentOption(val)
//         props.onSelect && props.onSelect(val)
//     }
//     return <div className={"centerH gapTop rwSelector"}>{props.txt || "Current Reader/Writer"}<select value={state.selected || defaultOpt} onChange={onSelect}>
//         {[defaultOpt, ...state.options].map(function (name, i) {
//             return <option value={name} key={i}>{name}</option>
//         })}
//     </select></div>
// }