'use client'
import * as React from "react";
import {SyntheticEvent} from "react";
import {
    Actions,
    ActionTypes,
    useRfidReaderContext
} from "@/app/components/formSubcomponents/readerWriterButtons/readerOptsContext";
import {Makeid} from "@/app/components/TopBar";
import {OnViewCreatorQuadCol} from "@/app/components/formSubcomponents/shared";
import {WriteRfidTag} from "@/app/components/serverActions";
import {Subform} from "@/app/components/common";


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

export function WriteTagFunc(dispatch: React.Dispatch<Actions>, id: string, selectedReader?: string): Promise<string> {
    return new Promise((resolve, reject) => {
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
        if (readerName === "goodTestRfid"){ // TODO: comment out
            return resolve(id)
        } else if (readerName === "" || readerName === "none" || readerName === "badTestRfid"){
            return reject("invalid reader name")
        } else {
            WriteRfidTag(id, selectedReader).then(()=>{
                dispatch({
                    type: ActionTypes.SET_LAST_READ_TAG,
                    payload: id,
                })
                dispatch({
                    type: ActionTypes.CLEAR_ERROR,
                })
                resolve(id)
            }).catch(e=>{
                let errTxt = "failed to write tag: "+JSON.stringify(e)
                console.error(errTxt);
                dispatch({
                    type: ActionTypes.SET_ERROR,
                    payload: errTxt,
                })
                reject(errTxt)
            })
        }

        resolve(id)
    })
}

export function ReadTagFunc(dispatch: React.Dispatch<Actions>, sess?: string, selectedReader?: string): Promise<string> {
    return new Promise((resolve, reject) => {
        // TODO: fix sess
        // ReadRfidTag(session, dispatch, selectedReader).then((id)=>{
        //     resolve(id)
        // },(err)=>{
        //     reject(err)
        // })
        // TODO: remove all below when reenabled!!!!!!!!!!
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
        if (readerName === "goodTestRfid"){ // TODO: comment out
            tagVal = "4Wj8HxCMmcs" // TODO: Test empty plate id
        } else if (readerName === "" || readerName === "none"){
            return
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

// TODO: TEST HEAVILY!
export function WriteRfidOvcArea(id:string):OnViewCreatorQuadCol{
    return {
        txt: "Write tag (dangerous)",
        newCreationArea: onCreate => <WriteRFIDArea id={id}
            handleTagWritten={(idWritten:string)=>{
                onCreate([{typeText: "Wrote Tag", node: <text>{idWritten}</text>}], true)
            }}/>,
    }
}

export function WriteRFIDArea( // TODO: use this on each page that has writeable ids!
    {
        handleTagWritten, id,
    }:{
        id:string,
        handleTagWritten:(id: string)=>void
    }) {
    const [locked, setLocked] = React.useState(true)
    const {state, dispatch} = useRfidReaderContext()
    const [writer, setWriter] = React.useState(state.selected)
    const writeTag = (e: React.MouseEvent)=>{
        console.log("attempting to write tag "+id+" to writer "+(writer || "none selected")) // TODO: DEL!
        e.preventDefault();
        e.stopPropagation();
        WriteTagFunc(dispatch, id, state.selected)
            .then(handleTagWritten)
            .catch(e=>console.error("got BAD tag write result: "+JSON.stringify(e))) // TODO: DEL!
    }
    return <Subform>
        <div>{"Setup Tag Writing"}</div>
        <div className={"inlineChildren"}>
            <text className={"mr-2"}>{"Unlocked"}</text>
            <input type="checkbox" checked={!locked} onChange={()=>{setLocked(!locked)}}/>
        </div>
        <ReaderWriterSelector txt={"Write to:"} onSelect={setWriter}/>
        {(writer && !locked) && <button className={"basicButtonSmall"} disabled={locked || !state.selected} onClick={writeTag}>
            {"Write "+id+" to writer: "+(writer || "none selected")}
        </button>}
    </Subform>
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