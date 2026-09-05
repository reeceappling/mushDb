'use client'
import * as React from "react";
import {SyntheticEvent} from "react";
import {
    Actions,
    ActionTypes,
    useRfidReaderContext
} from "@/app/components/formSubcomponents/readerWriterButtons/readerOptsContext";
import {OnViewCreatorQuadCol} from "@/app/components/formSubcomponents/shared";
import {Subform} from "@/app/components/common";
import {BaseExternalUrl} from "@/app/components/Constants";
import {ConfirmOrCancel} from "@/app/components/formSubcomponents/moveOnceUsed";
import TextBox from "@/app/components/formSubcomponents/textbox";
import {ReadTagButton, UseLatestReadTagButton} from "@/app/components/TopBar";


export interface rfidSelectorProps {
    defaultOption?: string,
    txt?: string,
    onSelect?: (s?: string) => void,
    headerLevel?: number
}

export default function ReaderWriterSelector(props:rfidSelectorProps) {
    const {state, dispatch} = useRfidReaderContext()
    const defaultOpt = props.defaultOption || "none"
    const onSelect = (e: SyntheticEvent<HTMLSelectElement, Event>) => {
        const val = e.currentTarget.value
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
            const toWrite = "no rfid reader selected"
            dispatch({
                type: ActionTypes.SET_ERROR,
                payload: toWrite,
            })
            reject(toWrite)
            return
        }
        const readerName = selectedReader
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
                const errTxt = "failed to write tag: "+JSON.stringify(e)
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
        if (!selectedReader) {
            const toWrite = "no rfid reader selected"
            dispatch({
                type: ActionTypes.SET_ERROR,
                payload: toWrite,
            })
            reject(toWrite)
            return
        }
        const readerName = selectedReader
        console.log("current reader name: "+readerName)// TODO: del!
        if (readerName === "goodTestRfid"){ // TODO: comment out
            const tagVal = "4Wj8HxCMmcs" // TODO: Test empty plate id
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
            return
        } else if (readerName === "" || readerName === "none"){
            reject("no reader selected...")
            return
        }
        ReadRfidTag(readerName).then((id)=>{
            // console.log("got tag id: "+id) // TODO: del
            dispatch({
                type: ActionTypes.SET_LAST_READ_TAG,
                payload: id,
            })
            dispatch({
                type: ActionTypes.SET_LAST_READER,
                payload: readerName,
            })
            dispatch({
                type: ActionTypes.CLEAR_ERROR,
            })
            resolve(id)
        },(err)=>{
            console.log("failed to read tag id: "+JSON.stringify(err)) // TODO: del
            reject(err)
        })
    })
}

export function ClearTagFunc(dispatch: React.Dispatch<Actions>, readerName?: string): Promise<string> {
    return new Promise((resolve, reject) => {
        if (!readerName) {
            const toWrite = "no rfid reader selected"
            dispatch({
                type: ActionTypes.SET_ERROR,
                payload: toWrite,
            })
            reject(toWrite)
            return
        }
        if (readerName === "goodTestRfid") { // TODO: comment out
            return resolve("Cleared")
        } else if (readerName === "badTestRfid") {
            return reject("failed to clear tag")
        } else if (readerName === "" || readerName === "none"){
            return reject("empty or 'none' writer name provided")
        } else {
            ClearRfidTag(readerName).then((responseString)=>{
                dispatch({
                    type: ActionTypes.SET_LAST_READ_TAG, // TODO: validate ok
                })
                dispatch({
                    type: ActionTypes.CLEAR_ERROR,
                })
                resolve(responseString)
            }).catch(e=>{
                const errTxt = "failed to write tag: "+JSON.stringify(e)
                console.error(errTxt); // TODO: del?
                dispatch({
                    type: ActionTypes.SET_ERROR,
                    payload: errTxt,
                })
                reject(errTxt)
            })
        }
    })
}


export function ReadRfidTag(readerName?: string):Promise<string> { // TODO: USE ME!!!
    if (!readerName) {
        throw "NO RFID READER SELECTED!"
    }
    return fetch(BaseExternalUrl + '/rfid/read/' + readerName, {
        method: 'GET',
        headers: {
            credentials: 'include',
            'Accept': 'text/html',
            'Content-Type': 'text/html'
        },
    }).then(resp=>{
        if (resp.status != 200) {
            throw "Error reading tag. Response status code " + resp.status
        }
        const contentType = resp.headers.get('Content-Type')
        if (contentType == null) {
            throw "Response had no content type!"
        }
        if (contentType != 'text/html') {
            throw "Unexpected response content type!"
        }
        return resp.text()
    })
}

export async function WriteRfidTag(toWrite: string, writerName: string) { // TODO: USE ME! NEEDS sessionInfo
    const resp = await fetch(BaseExternalUrl + '/rfid/write/' + writerName, {
        method: 'POST',
        headers: {
            credentials: 'include',
            'Accept': 'text/html',
            'Content-Type': 'application/json',
        },
        body: toWrite
    })
    if (resp.status != 200) {
        throw "Error reading tag. Response status (" + resp.status + ")" + resp.statusText
    }
    const contentType = resp.headers.get('Content-Type')
    if (contentType == null) {
        throw "Response had no content type!"
    }
    if (contentType != 'text/html') {
        throw "Unexpected response content type! " + contentType + " should be text/html"
    }
    return await resp.text()
}
export async function ClearRfidTag(writerName: string) {
    const resp = await fetch(BaseExternalUrl + '/rfid/clear/' + writerName, {
        method: 'DELETE',
        headers: {
            credentials: 'include',
            'Accept': 'text/html',
        },
    })
    if (resp.status != 200) {
        throw "Error reading tag. Response status (" + resp.status + ")" + resp.statusText
    }
    const contentType = resp.headers.get('Content-Type')
    if (contentType == null) {
        throw "Response had no content type!"
    }
    if (contentType != 'text/html') {
        throw "Unexpected response content type! " + contentType + " should be text/html"
    }
    return await resp.text()
}

// TODO: writeRfidTag!

export type selectReaderResult = {
    didRead: boolean;
    payload?: string;
};

export function SelectReaderFunc(dispatch: React.Dispatch<Actions>, doRead: boolean, session?: string, reader?: string): Promise<selectReaderResult> {
    const out:selectReaderResult = {didRead: doRead}
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

export function ClearRFIDButton(
    {
        props
    }:{
        props: {
            txt?:string,
            handleTagClear:()=>void
            handleTagClearError:(e:string)=>void
            handleTagClearCancel:()=>void
            session?: string
        }
    }) {
    const {state, dispatch} = useRfidReaderContext()

    return <button className={"removeButtonSmall"/* TODO: OK?*/}
                   onClick={()=>{
                       ConfirmOrCancel({
                           txt: "Are you sure you want to clear the tag on the "+state.selected+" reader/writer?",
                           onConfirm: ()=>{
                               ClearTagFunc(dispatch, state.selected)
                                   .then(props.handleTagClear)
                                   .catch(e=>{
                                       props.handleTagClearError("failed to clear tag on reader "+state.selected+": "+JSON.stringify(e))
                                   })
                           },
                           onCancel: props.handleTagClearCancel
                       })
                   }}>{props.txt || "Clear tag on reader (dangerous)"}</button>
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

export function IdInput({initial}:{initial?:string}){
    //const [err, setErr] = React.useState<string | undefined>(undefined)
    const [id, setId] = React.useState(initial)
    const {state} = useRfidReaderContext()
    return <div>
    <div>{"Main Collection Item By ID"}</div>
    <TextBox readonly={false} label={"ID"} value={id || ""} fieldName={"idInput"}
             updateTextHandler={setId}/>
    <ReadTagButton onResult={setId}/>
    {state.lastReadTag && <UseLatestReadTagButton onClick={(v) => {
        v && setId(v)
    }}/>}
</div>
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