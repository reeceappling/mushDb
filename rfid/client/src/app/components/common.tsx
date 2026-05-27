'use client'

import {defaultHeaderLevel} from "@/app/components/formSubcomponents/utils/headers";
import * as React from "react";
import {JSX, ReactNode, SetStateAction, SyntheticEvent, useEffect, useRef, useState} from "react";
import {
    Contamination,
    ContaminationForm,
    NewContaminationForm
} from "@/app/components/formSubcomponents/contaminations";
import EntryLink, {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {NumberToDate} from "@/app/components/formSubcomponents/date";
import {SplitAllEntries} from "@/app/components/formSubcomponents/shared";
import {NewPicWithNotesForm, PicWithNotesForm} from "@/app/components/formSubcomponents/picWithNotes";
import {BaseExternalUrl} from "@/app/components/Constants";
import ReaderWriterSelector, {
    ReadTagFunc,
} from "@/app/components/formSubcomponents/readerWriterButtons/readerSelector";
import {useRfidReaderContext} from "@/app/components/formSubcomponents/readerWriterButtons/readerOptsContext";
import {
    AssertDualListResult,
    AssertSubstrateRecipe,
    validatorForAssertion
} from "@/app/components/substrateRecipeClient";
import TestAndValidate from "@/app/components/testing/untested";
import {InputTextInlineTitle} from "@/app/components/formSubcomponents/numericInput";
import {AssertAgarRecipe} from "@/app/components/agarRecipeClient";
import {AssertAgarBatch} from "@/app/components/agarBatchClient";
import {AssertBag} from "@/app/components/bagClient";
import {AssertFruit} from "@/app/components/fruitClient";
import {AssertFruitingChamber} from "@/app/components/fruitingChamberClient";
import {AssertGrainBatch} from "@/app/components/grainBatchClient";
import {AssertJar} from "@/app/components/jarClient";
import {AssertJarRecipe} from "@/app/components/jarRecipeClient";
import {AssertLcRecipe} from "@/app/components/lcRecipeClient";
import {AssertLc} from "@/app/components/lcClient";
import {AssertLcSyringe} from "@/app/components/lcSyringeClient";
import {AssertMss} from "@/app/components/mssClient";
import {AssertPcRun} from "@/app/components/pcRunClient";
import {AssertPlate} from "@/app/components/plateClient";
import {AssertProject} from "@/app/components/projectClient";
import {AssertSale} from "@/app/components/saleClient";
import {AssertSlant} from "@/app/components/slantClient";
import {AssertSpecies} from "@/app/components/speciesClient";
import {AssertSporePrint} from "@/app/components/sporePrintClient";
import {AssertSporeSwab} from "@/app/components/sporeSwabClient";
import {AssertStasisTube} from "@/app/components/stasisTubeClient";
import {AssertSubspecies} from "@/app/components/subspeciesClient";
import {AssertSubstrateBatch} from "@/app/components/substrateBatchClient";
import {AssertUser} from "./userClient";
import {AssertWaterJar} from "@/app/components/waterJarClient";
import {AssertTransfer} from "@/app/components/transferClient";
import SpeechRecognition, {useSpeechRecognition} from "react-speech-recognition";
import {ActionTypes, useDictationContext} from "@/app/components/formSubcomponents/dictationContext/dictationContext";


export function SendMultipartRequest(url: string, cookies: string, formData: FormData) {
    return fetch(url, {
        method: 'Post',
        body: formData,
        credentials: 'include',
        headers: {
            credentials: 'include',
            'Cookie': cookies, // TODO: does this need to be here? I think so for multipart
            'Access-Control-Allow-Origin': '*',
        },
    })
}

export function SayString(toDictate: string) {
    DictateString(toDictate)
}

export function DictateString(toDictate: string) { // TODO: USE!
    if ('speechSynthesis' in window) {
        window.speechSynthesis.speak(new SpeechSynthesisUtterance(toDictate))
    } else {
        throw "client speech synthesis not currently available"
    }
}

// TODO: DICTAPHONES SHOULD BE USED IN:
// TODO: creates: anything that needs a sterile environment (LIST)
// TODO: views: all of them!
// TODO: consider embedding dictaphones in notes areas for views and creates, and controlling the notes with a context of some sort?
export function Dictaphone({createNoteHandler}: { createNoteHandler?: (note: string) => void }) {
    // const cmds = ["simon says", "new note"]
    const timeoutRef = useRef<NodeJS.Timeout | undefined>(undefined)
    const [activeCommand, setActiveCommand] = useState<string | undefined>(undefined)
    const [startedBody, setStartedBody] = useState(false)
    const listenArgs = {
        continuous: true, // TODO: ok? was false
        interimResults: true, // TODO: ok? was false
        language: "en-US",
    }

    // const startBodyListener = ()=>{
    //
    // }
    // const startCommandListener = ()=>{
    //
    // }
    //
    // //const fullCmdRegex = new RegExp("(?<=^command )simon says [a-zA-Z0-9 ]+(?= end dictation)")
    // //const startDictationString = "command"
    // const resetString = "clear dictation"
    // const resetDictationRegex = new RegExp(resetString, "g")
    // const endBodyString = "end dictation"
    // const endDictationRegex = new RegExp("^[a-zA-Z0-9 ]+ "+endBodyString+"$)", "g")
    // const bodyCommand = "* "+endBodyString
    // const simonSaysRegex = regexForCmd("simon says")
    // const cmdRegex = [simonSaysRegex]
    // const removePrefix = (str: string, pre: string):string => {
    //     str.slice(pre.length);
    // }
    // const bodyCallback = (command: string, resetTranscript:()=>void):void=>{
    //     const body = command.substring(0,command.length-(2+endBodyString.length)) // TODO: ensure length right
    //     switch(activeCommand){
    //         case undefined:
    //             // TODO: ERROR
    //     }
    // }
    // const cmdCallback = (command: string, resetTranscript:()=>void):void => {
    //     const commandAndBody = removePrefix(lessEnd, prefixes[0])
    //     switch(command){
    //         case cmds[0]: //simon says
    //             setActiveCommand(cmds[0])
    //             break;
    //         default:
    //     }
    //     if (lessEnd.startsWith(prefixes[0])){
    //         let body = removePrefix(lessEnd, prefixes[0])
    //
    //     }
    //     resetTranscript()
    // }
    const commands = [
        {
            command: ["reset dictation", "clear transcript", "reset transcript"],
            callback: () => {
                resetTranscript()
                setActiveCommand(undefined)
            },
            matchInterim: true,
        },
        {
            command: ["repeat after me", "simon says"],
            callback: () => {
                resetTranscript()
                setActiveCommand("repeat after me")
            },
            matchInterim: true,
        },
        {
            command: [
                "new note",
                "create note",
                "create new note",
                "create a note",
                "create a new note",
                "make note",
                "make a note",
                "make new note",
                "make a new note",

            ],
            callback: () => {
                resetTranscript()
                setActiveCommand("create note")
            },
            matchInterim: true,
        },
    ]
    const {transcript, listening, resetTranscript, browserSupportsSpeechRecognition} = useSpeechRecognition({
        commands: commands,
    });
    // 3-Second Timeout Logic
    useEffect(() => {
        // Clear existing timeout each time a new transcript word is detected
        if (timeoutRef.current) {
            clearTimeout(timeoutRef.current);
        }

        // Set a new 3-second timer
        const currentText = transcript
        // TODO: handle 0-length transcripts?
        const onTimeout = () => {
            switch (activeCommand) {
                case "repeat after me":
                    console.log("repeat after me: " + currentText)
                    SayString(currentText)
                    break;
                // TODO: CREATE PLATE? Bag, Slant, Transfer?
                case "create note":
                    // TODO: repeat and ask to save??????
                    console.log("created note: " + currentText)
                    createNoteHandler && createNoteHandler(currentText)
                    break;
                default:
                    return
                // TODO: this!
            }
            setActiveCommand(undefined)
            resetTranscript()
        }
        timeoutRef.current = setTimeout(onTimeout, 3000);

        return () => clearTimeout(timeoutRef.current);
    }, [transcript, activeCommand]);

    if (!browserSupportsSpeechRecognition) {
        return <span>{"Browser doesn't support speech recognition."}</span>;
    }

    return (
        <div>
            <p>{"Microphone: " + (listening ? 'on' : 'off')}</p>
            <button onClick={e => {
                e.stopPropagation();
                SpeechRecognition.startListening(listenArgs)
            }}>{"Start"}</button>
            <button onClick={e => {
                e.stopPropagation();
                SpeechRecognition.stopListening()
            }}>{"Stop"}</button>
            <button onClick={e => {
                e.stopPropagation();
                resetTranscript()
            }}>Reset
            </button>
            <p>{transcript}</p>
        </div>
    );
};

// TODO: USE ON TFID VIEW PAGES!
// TODO: SHOULD ADD WHERE NEEDED
// TODO: LIKELY NEEDS MAJOR OVERHAUL
export function ViewPageDictaphone({doUpdate}: {
    doUpdate: () => void
}) {
    const rfidRdr = useRfidReaderContext()
    const dict = useDictationContext()
    // TODO: let readerWriter = state.selected // TODO: or lastReaderUsed???
    const listenArgs = {
        continuous: true,
        interimResults: true,
        language: "en-US",
    }
    const handleViewById = (idToSearch: string) => {
        getPathFor(idToSearch).then((path) => {
            location.assign(BaseExternalUrl + "/view/" + path)
        }).catch((err) => {
            console.log("failed to get path for id: " + JSON.stringify(err))
            SpeechRecognition.startListening(listenArgs)
        })
    }
    const commands = [
        {
            command: ["create transfer", "new transfer"],
            callback: () => {
                resetTranscript()
                SpeechRecognition.stopListening()
                dict.dispatch({type: ActionTypes.SET_CURRENT,payload:"create transfer"})
            },
            matchInterim: true,
        },
        {
            command: ["view tag"], // TODO: ok?
            callback: () => {
                SpeechRecognition.stopListening()
                resetTranscript()
                ReadTagFunc(rfidRdr.dispatch, undefined, rfidRdr.state.selected)
                    .then(handleViewById)// redir to the new page
                    .catch(e => {
                        console.error("failed to read linking tag: " + JSON.stringify(e))
                        SpeechRecognition.startListening(listenArgs)
                    })
            },
            matchInterim: true,
        },
        {
            command: ["submit updates"], // TODO: ok?
            callback: () => {
                SpeechRecognition.stopListening()
                resetTranscript()
                doUpdate()
                SpeechRecognition.startListening(listenArgs)
            },
            matchInterim: true,
        },
        {
            command: [
                "new note",
                "create note",
                "create new note",
                "create a note",
                "create a new note",
                "make note",
                "make a note",
                "make new note",
                "make a new note",
                "add a new note",
                "add new note",
                "add a note",
                "add note",
            ],
            callback: () => {
                SpeechRecognition.stopListening()
                resetTranscript()
                dict.dispatch({type: ActionTypes.SET_CURRENT,payload:"create note"})
            },
            matchInterim: true,
        },
    ]
    const {transcript, listening, resetTranscript, browserSupportsSpeechRecognition} = useSpeechRecognition({
        commands: commands,
    });

    if (!browserSupportsSpeechRecognition) {
        return <span>{"Browser doesn't support speech recognition."}</span>;
    }
    useEffect(() => { // TODO: validate works right
        if (dict.state.current === "main") {
            SpeechRecognition.startListening(listenArgs)
        }
    }, [dict.state.current])

    return (
        <div>
            <button onClick={e => {
                e.stopPropagation();
                SpeechRecognition.startListening(listenArgs)
            }}>{"Enable Dictation"}</button>{/* TODO: dictation enablement in cookies? We want to be able to traverse pages without touching the screen*/}
            <button onClick={e => {
                e.stopPropagation();
                SpeechRecognition.stopListening()
            }}>{"Disable Dictation"}</button>
        </div>
    );
};

export function AddNoteDictaphone({parent,createNote}:{parent?:string,createNote:(s:string)=>void}){
    // Always created in a state that is not listening by default
    try {
        const {state, dispatch} = useDictationContext()
        const listenArgs = {
            continuous: false,
            interimResults: false, // TODO: UNSURE IF WE WANT THIS OR NOT
            language: "en-US",
        }
        const commands = [
            {
                command: ["* complete note"],
                callback: (note: string) => {
                    SpeechRecognition.stopListening()
                    createNote(note)
                    resetTranscript()
                    dispatch({type: ActionTypes.SET_CURRENT,payload:parent||"main"}) // Because if this is not right below the main parent, then it should revert to the closest parent
                },
                matchInterim: true,
            },
        ]
        const {transcript, listening, resetTranscript, browserSupportsSpeechRecognition} = useSpeechRecognition({
            commands: commands,
        });
        const parentPrefix = ((parent && parent !== "main")?parent+".":"")
        useEffect(() => { // TODO: validate works right
            if (state.current === parentPrefix+"create note") {
                SpeechRecognition.startListening(listenArgs)
            }
        }, [state.current])
    } catch (e){
        console.error("failed to create note dictation component: " + JSON.stringify(e))
        return null
    }
}

// TODO: USE THIS!
export function CreateTransferDictaphone({submit,deleteLastNote,setDstId,setTransferReason}:{
    submit:()=>void,
    deleteLastNote:()=>void,
    setDstId:(id:string)=>void,
    setTransferReason:(id:string)=>void,
}){
    // Always created in a state that is not listening by default
    try {
        const rfidCtx = useRfidReaderContext()
        const {state, dispatch} = useDictationContext()
        const listenArgs = {
            continuous: false,
            interimResults: false, // TODO: UNSURE IF WE WANT THIS OR NOT
            language: "en-US",
        }
        const commands = [
            {
                command: ["scan destination"],
                callback: () => {
                    SpeechRecognition.stopListening()
                    resetTranscript()
                    ReadTagFunc(rfidCtx.dispatch, undefined, rfidCtx.state.selected)
                        .then((idRead)=>{
                            setDstId(idRead) // TODO: validate working
                            SpeechRecognition.startListening(listenArgs)
                        })
                        .catch(e => {
                            console.error("failed to read linking tag: " + JSON.stringify(e))
                            SpeechRecognition.startListening(listenArgs)
                        })
                },
                matchInterim: true,
            },
            {
                command: ["* is the transfer reason"], // TODO: EW!
                callback: (arg:string) => {
                    SpeechRecognition.stopListening()
                    resetTranscript()
                    setTransferReason(arg) // TODO: validate working
                    SpeechRecognition.startListening(listenArgs)
                },
                matchInterim: true,
            },
            {
                command: ["list transfer reason options"], // TODO: EW!
                callback: () => {
                    // TODO: THIS!
                },
                matchInterim: true,
            },
            // TODO: add notes (change to "create transfer.create note" in dictation context)
            {
                command: ["delete last note"],
                callback: () => {
                    SpeechRecognition.stopListening()
                    resetTranscript()
                    deleteLastNote()// TODO: THIS!
                    SpeechRecognition.startListening(listenArgs)
                },
                matchInterim: true,
            },
            { // TODO: "with note * submit transfer" ?
                command: ["submit current transfer"],
                callback: () => {
                    SpeechRecognition.stopListening()
                    resetTranscript()
                    submit()
                    dispatch({type: ActionTypes.SET_CURRENT, payload:"main"}) // main is parent of transfer
                },
                matchInterim: true,
            },
        ]
        const {transcript, listening, resetTranscript, browserSupportsSpeechRecognition} = useSpeechRecognition({
            commands: commands,
        });
        useEffect(() => { // TODO: validate works right
            if (state.current === "create transfer") {
                SpeechRecognition.startListening(listenArgs)
            }
        }, [state.current])
    } catch (e){
        console.error("failed to create transfer dictation component: " + JSON.stringify(e))
        return null
    }
}

// TODO: USE THIS!
export function MainCollectionInputOrRead({label, onIdSelected, copyText}: {
    label?: string,
    onIdSelected: (s: string) => void
    copyText?: string,
}) {
    const {state, dispatch} = useRfidReaderContext()
    const [id, setId] = useState<string>("");
    const updateId = (newId: string) => {
        setId(newId)
        onIdSelected(newId)
    }
    return <div>
        {/* INPUT FOR MAINCOLLECTIONID */}
        <InputTextInlineTitle label={"ID TO:"} value={id} readonly={false} errorMessage={undefined/* TODO: ???*/}
                              placeholder={"Destination"} onChange={(s) => updateId(s || "")}/>
        {/*<TextBox label={label || "Main Collection Id Input: "} value={id} fieldName={"mainCollIdInput"}*/}
        {/*         updateTextHandler={updateId} readonly={false}/>*/}
        {/* BUTTON TO READ MAIN COLL ID */}
        <ReaderWriterSelector txt={"select rfid reader"} onSelect={(wr) => { // TODO: wr ok here or state.selected?
            ReadTagFunc(dispatch, undefined, wr).then(updateId)
        }}/>
        {/*<RfidSelectorWithReadButton handleTagRead={updateId} readButtonTxt={"read from current tag reader"}*/}
        {/*                            readerWriterTxt={"select rfid reader"} onWriterSelect={(wr)=>{*/}
        {/*    ReadTagFunc(dispatch, undefined, state.selected).then(updateId)*/}
        {/*}}/>*/}
        {/* BUTTON TO USE LAST READ ID */}
        {state.lastReadTag !== undefined && <div>
            <button className={"basicButton"} onClick={() => {
                state.lastReadTag && setId(state.lastReadTag)
            }}>{copyText || "Copy last read id"}</button>
        </div>}
    </div>
}

export function AssertArrayResult<T>(input: any, validateEntry: (inp: any) => void): asserts input is T[] {
    if (!Array.isArray(input)) {
        throw new Error('not an array');
    }
    try {
        if (!CheckArrayType(input, validatorForAssertion(validateEntry))) {
            throw new Error('incorrect item types in array');
        }
    } catch (e) {
        throw e;
    }

    return
}

export function CheckArrayType<T>(arr: T[], typeChecker: (item: T) => boolean): boolean {
    return arr.every(typeChecker);
}

export function RequiredKey(key: string, input: any, validateType: (val: any) => boolean): boolean {
    return key in input && validateType(input[key])
}

// // TODO: Redundant get rid of?
// export function RequiredSimpleKey(key: string, input: any, expType: string): boolean {
//     return RequiredKey(key, input, (val: any) => {
//         return typeof val === expType
//     })
// }

export function OptionalKey(key: string, input: any, validateIfExists: (inp: any) => boolean): boolean {
    return (key in input) ? validateIfExists(input[key]) : true
}

export function OptionalSimpleKey(key: string, input: any, expType: string): boolean {
    return OptionalKey(key, input, IsType(expType))
}

export function IsType(finalType: string): (inpt: any) => boolean {
    return (inp: any) => {
        return typeof inp === finalType
    }
}

export function RequiredArrayOfType(key: string, input: any, validateChildren: (child: any) => boolean): boolean {
    return RequiredKey(key, input, (val: any): boolean => {
        return Array.isArray(val) && CheckArrayType(val, validateChildren)
    })
}

export function OptionalArrayOfType(key: string, input: any, validateChildren: (child: any) => boolean): boolean {
    return OptionalKey(key, input, (chd: any): boolean => {
        return Array.isArray(chd) && CheckArrayType(chd, validateChildren)
    })
}

// TODO: delete if unneeded
// export function OptionalMapOfType(key: string, input: any, validateChildren: (child: any) => boolean): boolean {
//     if (typeof input !== 'object') {
//         throw new Error('Input is not an object! Input is ' + typeof input);
//     }
//     return OptionalKey(key, input, (chd: any): boolean => {
//         return CheckArrayType(chd, validateChildren)
//     })
// }

export function ViewInNewTabButton({entryType, id}: { entryType: string, id: string }) {
    return <EntryLinkWrapper props={{linkId: encodeURI(encodeURI(id)), entryType: entryType, openInNewTab: true}}>
        <button className={"basicButtonSmall"}>{"View"}</button>
    </EntryLinkWrapper>
}

export function ListItemsRequest(entryType: string) {
    return fetch(BaseExternalUrl + "/db/list/" + entryType, {
        method: 'Get',
        credentials: 'include',
        headers: {
            credentials: 'include',
            'Accept': 'application/json',
        },
    }).then((res) => {
        if (!res.ok) {
            throw new Error('response not ok. Status=' + res.status + ', body=' + res.text())
        }
        return res.json().then(result => {
            let asserter: (x: any) => void = () => false
            switch (entryType) {
                case "agarBatches":
                    asserter = AssertAgarBatch;
                    break;
                case "agarRecipes":
                    asserter = AssertAgarRecipe;
                    break;
                case "bags":
                    asserter = AssertBag;
                    break;
                case "fruits":
                    asserter = AssertFruit;
                    break;
                case "fruitingChambers":
                    asserter = AssertFruitingChamber;
                    break;
                case "grainBatches":
                    asserter = AssertGrainBatch;
                    break;
                case "jars":
                    asserter = AssertJar;
                    break;
                case "jarRecipes":
                    asserter = AssertJarRecipe;
                    break;
                case "lcs":
                    asserter = AssertLc;
                    break;
                case "lcRecipes":
                    asserter = AssertLcRecipe;
                    break;
                case "lcSyringes":
                    asserter = AssertLcSyringe;
                    break;
                case "mss":
                    asserter = AssertMss;
                    break;
                case "pcRuns":
                    asserter = AssertPcRun;
                    break;
                case "plates":
                    asserter = AssertPlate;
                    break;
                case "projects":
                    asserter = AssertProject;
                    break;
                case "sales":
                    asserter = AssertSale;
                    break;
                case "slants":
                    asserter = AssertSlant;
                    break;
                case "species":
                    asserter = AssertSpecies;
                    break;
                case "sporePrints":
                    asserter = AssertSporePrint;
                    break;
                case "sporeSwabs":
                    asserter = AssertSporeSwab;
                    break;
                case "stasisTubes":
                    asserter = AssertStasisTube;
                    break;
                case "subspecies":
                    asserter = AssertSubspecies;
                    break;
                case "substrateBatches":
                    asserter = AssertSubstrateBatch;
                    break;
                case "substrateRecipes":
                    asserter = AssertSubstrateRecipe;
                    break;
                case "transfers":
                    asserter = AssertTransfer;
                    break;
                case "users":
                    asserter = AssertUser;
                    break;
                case "waterJars":
                    asserter = AssertWaterJar;
                    break;
                default:
                    throw new Error("invalid type but got response. Should never happen");
                    break;
            }
            switch (entryType) {
                case "agarRecipes":
                case "jarRecipes":
                case "lcRecipes":
                case "substrateRecipes":
                    AssertDualListResult(result, asserter);
                    break;
                default:
                    AssertArrayResult(result, asserter);
                    break;
            }
            return result
        })
    })
}

export function IsString(item: any): boolean {
    return typeof item === 'string'
}

export function IsBool(item: any): boolean {
    return typeof item === 'boolean'
}

export function HeaderLevel(lvl?: number) {
    return lvl || defaultHeaderLevel
}

export interface ListPageItems<T> {
    data: T[],
    onClick?: (v: T) => void
    withLink?: boolean,
}

export interface InlineProps<T> {
    data: T,
    expandByDefault?: boolean,
    onClick?: (v?: T) => void
    headerLevel?: number
    idIsLink?: boolean
    showMainPageButton?: boolean
}

export interface SingleListProps<T> {
    data: T[],
    onClick: (v: T) => void
}

export interface TwoListProps<T> {
    recent: T[],
    standard: T[],
    onClick: (v: T) => void
}

export function InlineSubArea(
    {
        props, children
    }: {
        props: {
            className?: string
        },
        children: ReactNode,
    }) {
    return <div data-cy-id="InlineSubAreaWrapper" className={props.className}>
        <div data-cy-id="InlineSubArea" className={"inlineSubArea"}>
            {children}
        </div>
    </div>
}

export function InlineExpansionArea(
    {
        props, children
    }: {
        props: {
            expanded?: boolean
        },
        children: ReactNode,
    }) {
    if (!props.expanded) {
        return null
    }
    return <InlineSubArea data-cy-id="InlineExpansionArea" props={{}}>
        {children}
    </InlineSubArea>
}

export function InlineExpansionButton(
    {
        setExpanded, expanded
    }: {
        setExpanded: (value: SetStateAction<boolean | undefined>) => void
        expanded?: boolean
    }) {
    return <div data-cy-id="InlineExpansionButtonWrapper">
        <button className={"basicButton"} data-cy-id="InlineExpansionButton" onClick={(e) => {
            e.stopPropagation();
            setExpanded(!expanded)
        }}>{expanded ? "See less" : "See more"}</button>
    </div>
}

export function TwoValuePlusUnknownSelector({pre, updateParent, initial, trueStr, falseStr, className}: {
    pre: string,
    updateParent?: (v?: boolean) => void,
    initial?: boolean,
    trueStr: string,
    falseStr: string
    className?: string
}) {
    if (initial !== undefined) {
        return <div>{pre + (initial ? trueStr : falseStr)}</div>
    }
    const strForBool = (s?: boolean) => {
        return ((s === undefined) ? "unknown" : (s ? trueStr : falseStr))
    }
    const [selected, setSelected] = useState<boolean | undefined>(initial)
    const boolForStr = (s: string) => {
        return ((s === "unknown") ? undefined : (s === trueStr))
    }

    const selectHandler = (e: SyntheticEvent<HTMLSelectElement, Event>) => {
        let val = boolForStr(e.currentTarget.value)
        updateParent && updateParent(val)
        setSelected(val)
    }
    return <div className={className}>
        <div>{pre}</div>
        <select className={"tailwindSelector"} value={strForBool(selected)} onChange={selectHandler}>
            <option value={"unknown"}>{"unknown"}</option>
            <option value={trueStr}>{trueStr}</option>
            <option value={falseStr}>{falseStr}</option>
        </select>
    </div>
}

export function ConfirmedCleanSelector(// TODO: validate works now via a test LC
    {updateParent, initial}:
    {
        updateParent: (b?: boolean) => void, initial?: boolean
    }) {
    return <TwoValuePlusUnknownSelector pre={"Confirmed Clean: "} updateParent={updateParent} initial={initial}
                                        trueStr={"clean"} falseStr={"contaminated"}/>

    // const strForBool = (s?: boolean) => {
    //     return ((s === undefined) ? "unknown" : (s ? "clean" : "contaminated"))
    // }
    // const [selected, setSelected] = useState<string>(strForBool(initial))
    // const boolForStr = (s: string) => {
    //     return ((s === "unknown") ? undefined : (s === "clean"))
    // }
    //
    // const selectHandler = (e: SyntheticEvent<HTMLSelectElement, Event>) => {
    //     let val = e.currentTarget.value
    //     selProps.doSelect(boolForStr(val))
    //     setSelected(val)
    // }
    // return <div className={"confirmedCleanSelector"}>{/* TODO: STYLING!!!!*/}
    //     <div>{"Confirmed Clean: "}</div>
    //     <select className={"tailwindSelector"} value={selected} onChange={selectHandler}>
    //         <option value={"unknown"}>{"unknown"}</option>
    //         <option value={"clean"}>{"clean"}</option>
    //         <option value={"contaminated"}>{"contaminated"}</option>
    //     </select>
    // </div>
}

export function YesNoSelector({pre, updateParent, initial, className}: {
    pre: string,
    updateParent?: (v?: boolean) => void,
    initial?: boolean
    className?: string
}) {
    return <TwoValuePlusUnknownSelector pre={pre} updateParent={updateParent} initial={initial} trueStr={"yes"}
                                        falseStr={"no"} className={className}/>
}

export function ConfirmedCleanArea(
    {
        readonly, initial, headerLevel, onSelect
    }: {
        readonly?: boolean
        initial?: boolean
        headerLevel?: number
        onSelect?: (c?: boolean) => void
    }) {
    return <YesNoSelector pre={"Confirmed Clean:"} initial={initial} updateParent={onSelect}/>
    // if (readonly) {
    //     return <div className={"confirmedCleanArea"}>
    //         <div>{"Confirmed Clean:"}</div>
    //         <div>{(initial === undefined) ? "Unknown" : (initial ? "Yes" : "No")}</div>
    //     </div>
    // }
    // return <div className={"confirmedCleanArea"}><ConfirmedCleanSelector initial={initial} updateParent={(v)=> {
    //     onSelect && onSelect(v)
    // }
    // }/></div>

}

export type DisplayInput = {
    id: string;
    readonly: boolean;
    data: any
    headerLevel?: number
    isTopLevel: boolean
    cookies: string
}

export type ImportDisplayInput = {
    headerLevel: number
    cookies: string
}

export function DisposedContamArea( // TODO: THIS AND USE THIS WHEN NEEDED!!!
    {
        headerLevel, disposed, contams
    }: {
        disposed?: number
        contams?: Contamination[]
        headerLevel?: number
    }) {
    return <div>
        <TestAndValidate todos={["DisposedContamArea NOT IMPLEMENTED!"]}>
            {"DisposedContamArea NOT IMPLEMENTED!"}
        </TestAndValidate>
    </div> // TODO: THIS!
}

export function DisposedSaleContamArea(
    {
        contams, sale, disposed, headerLevel
    }: {
        contams?: Contamination[]
        sale?: string
        disposed?: number
        headerLevel?: number
    }) {
    const sectionHeader = <div>{"Status: "}</div>
    if (sale) {
        const displayId = sale
        return <div>
            {sectionHeader}
            <div>{"Sold in sale "}
                <EntryLink props={{
                    displayedId: displayId,
                    linkId: displayId,
                    entryType: "sale",
                    openInNewTab: true
                }}>{displayId}</EntryLink>
            </div>
        </div>
    }
    let contamToUse: Contamination = {time: 0, confirmed: false, mold: false, bacteria: false, location: ""}
    if (contams !== undefined && contams.length === 0) {
        contamToUse.time = contams[contams.length - 1].time
        for (let i = 0; i < contams.length; i++) {
            if (!contamToUse.confirmed && contams[i].confirmed) {
                contamToUse.confirmed = true
            }
            if (!contamToUse.mold && contams[i].mold) {
                contamToUse.mold = true
            }
            if (!contamToUse.bacteria && contams[i].bacteria) {
                contamToUse.bacteria = true
            }
            if (contams[i].location) {
                contamToUse.location = contams[i].location
            }
        }
    }
    let contamLine: JSX.Element | null = null
    if (contamToUse.mold || contamToUse.bacteria) {
        let contamType = contamToUse.mold ? "mold" : "bacteria"
        if (contamToUse.mold && contamToUse.bacteria) {
            contamType = "mold, bacteria"
        }
        let lastContamPart = (" last cited " + NumberToDate(new Date(contamToUse.time)))
        contamLine = <div>
            <div>{(contamToUse.confirmed ? "Confirmed" : "Unconfirmed") + " contamination (" + contamType + ")" + lastContamPart}</div>
        </div>
    }
    let disposedSection = <div>
        {disposed ? "Disposed on " + NumberToDate(new Date(disposed)) : "Available"}{/* TODO: DIFFERENT STYLING BASED ON ANSWER?*/}
    </div>
    return <div>
        <div>{sectionHeader}</div>
        {contamLine}
        {disposedSection}
    </div>
}

// TODO: del if unused
// export function SaleAndDisposedArea({sale, disposed, headerLevel, readonly}: { // TODO: USE THIS WHERE NEEDED!!!!
//     sale?: string,
//     disposed?: number,
//     headerLevel?: number,
//     readonly: boolean
// }) {
//     if (sale) {
//         return <SaleArea sale={sale} readonly={true} headerLevel={headerLevel} canCreateSale={false}/>
//     }
// }

export interface NewEntryIdInput {
    headerLevel?: number,
    onCreate?: (id: string) => void
    redirectOnCreate: boolean
}


export interface NewEntryInput<T> {
    isTopLevel: boolean
    onCreate?: (newItem: T) => void
}

// TODO: MOVE THIS
export async function getTypeFor(id: string) { // TODO: ensure this works????
    // TODO: USE EXAMPLE ITEMS FOR DEV ENVIRONMENT!
    return await fetch(BaseExternalUrl + "/typeOf/" + id, {
        method: "GET",
        headers: {
            credentials: 'include',
            SessionId: "FIXME!!!", // TODO; THIS
        },
    }).then(HandleTxtResponse)
        .then((entryType) => {
            return entryType
        })
        .catch((error) => {
            throw error
        });
}

export async function getPathFor(id: string) { // TODO: ensure this works????
    let resp = await fetch(BaseExternalUrl + "/db/pathFor/" + id, {
        method: "GET",
        headers: {
            credentials: 'include',
            'Content-type': 'application/json'
        },
    })
    if (!resp.ok) {
        throw "failed to get path for id"
    }
    return await resp.text()
}

function webUrl(subPath: string) {
    return BaseExternalUrl + subPath
}

function apiUrl(subPath: string) {
    return BaseExternalUrl + "/db" + subPath
}

// TODO: use all of these all over the place!
export function viewUrlFor(itemType: string, newId: string) {
    return webUrl("/view/" + itemType + "/" + newId)
}

export function viewApiUrlFor(itemType: string, id: string) {
    return apiUrl("/get/" + itemType + "/" + id)
}

export function createUrlFor(itemType: string) {
    return webUrl("/create/" + itemType)
}

export function createApiUrlFor(itemType: string) {
    return apiUrl("/create/" + itemType)
}

export function importUrlFor(itemType: string) {
    return webUrl("/import/" + itemType)
}

export function importApiUrlFor(itemType: string) {
    return apiUrl("/import/" + itemType)
}

export function updateApiUrlFor(itemType: string, id: string) {
    return apiUrl("/update/" + itemType + "/" + id) // TODO: ensure ok
}


// TODO: fix inputs and use this everywhere we can???
export function CreateNewEntryButton(handler: { onSubmit: () => void }) {
    return <button className={"greenButton buttonFullWidth"} onClick={handler.onSubmit}>{"Create!"}</button>
}

export function resolvePicsFormData(picsIn: SplitAllEntries<PicWithNotesForm, NewPicWithNotesForm>) {
    let newImages: File[] = new Array(picsIn.new.length)
    let dataOut = {existing: picsIn.existing, new: new Array(picsIn.new.length)}
    for (let i = 0; i < picsIn.new.length; i++) {
        let toSend = picsIn.new[i]
        if (toSend.img === undefined) {
            throw new Error("new image " + i + " is undefined")
        } else {
            newImages[i] = toSend.img
        }
        dataOut.new[i] = {
            time: toSend.time,
            notes: toSend.notes.new.map(n => {
                return n.data
            })
        }
    }
    return {
        images: newImages,
        obj: dataOut,
    }
}

export function resolveContamsFormData(inp: SplitAllEntries<ContaminationForm, NewContaminationForm>) {
    let conts: (File | undefined)[] = new Array(inp.new.length)
    let dataOut = {existing: inp.existing, new: new Array(inp.new.length)}
    for (let i = 0; i < inp.new.length; i++) {
        dataOut.new[i] = {
            time: inp.new[i].time,
            confirmed: inp.new[i].confirmed,
            bacteria: inp.new[i].bacteria,
            mold: inp.new[i].mold,
            notes: inp.new[i].notes
        }
        conts[i] = inp.new[i].file
    }
    return {
        images: conts,
        obj: dataOut,
    }
}

export function setFormImages(formData: FormData, filePrefix: string, pics: any[]) {
    for (let i = 0; i < pics.length; i++) {
        const fileName = filePrefix + "-" + i
        if (pics[i] === undefined) {
            console.log("Picture undefined, " + fileName)
            continue
        }
        console.log("Picture set, " + fileName)
        formData.set(fileName, pics[i], fileName)
    }
}

export function setFormData(formData: FormData, dataObj: any) {
    formData.set("data", JSON.stringify(dataObj))
}

export function HandleJsonResponse(res: Response): Promise<any> {
    checkResponseStatus(res)
    return res.json()
}

export function HandleTxtResponse(res: Response): Promise<string> {
    checkResponseStatus(res)
    return res.text()
}

function checkResponseStatus(res: Response) {
    if (!res.ok || res.status !== 200) {
        throw "[(response status " + res.status + " " + res.statusText + ")]"
    }
    return
}