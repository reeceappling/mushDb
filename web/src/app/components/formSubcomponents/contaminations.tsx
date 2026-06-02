// non-client even though it uses state?

import {AllEntries, Data, InitialToAllEntries, SplitAllEntries} from "@/app/components/formSubcomponents/shared";
import {IsValidNote, Note, NotesFormArea} from "@/app/components/formSubcomponents/notes";
import {ImageLocationFor, PicWithNotesIncoming} from "@/app/components/formSubcomponents/picWithNotes";
import ImageSelector from "@/app/components/formSubcomponents/imageSelector";
import {OptionalArrayOfType, OptionalSimpleKey} from "@/app/components/common";
import DateArea from "@/app/components/formSubcomponents/date";
import {useContext, useEffect, useState} from "react";
import {DepthContext} from "@/app/components/formSubcomponents/depthContext/depth";
import {RemoveToggle} from "@/app/components/formSubcomponents/commonClient";
import TestAndValidate from "@/app/components/testing/untested";
import {InitialNotesState} from "@/app/components/formSubcomponents/initialState";

export const ExampleImageLocation: string = "test.jpg"


export interface Contamination {
    time: number
    confirmed: boolean,
    bacteria: boolean,
    mold: boolean,
    location: string
    notes?: Note[]
}

export interface ContaminationForm {
    time: number
    confirmed: boolean,
    bacteria: boolean,
    mold: boolean,
    location: string,
    notes: AllEntries<Note>
}

export interface NewContaminationForm {
    time: number,
    confirmed: boolean,
    bacteria: boolean,
    mold: boolean,
    file?: File,
    notes: Note[]
}

export function IsValidContamination(input: any): boolean {
    return typeof input === 'object' &&
        'time' in input && typeof input.time === 'number' &&
        'confirmed' in input && typeof input.confirmed === 'boolean' &&
        'bacteria' in input && typeof input.bacteria === 'boolean' &&
        'mold' in input && typeof input.mold === 'boolean' &&
        OptionalSimpleKey('location', input, "string") &&
        OptionalArrayOfType('notes', input, IsValidNote)
}

export function InitialContamState(contamination?: Contamination[]): SplitAllEntries<ContaminationForm, NewContaminationForm> {
    return {
        existing:
            contamination === undefined ? [] : contamination.map((ctn) => {
                let out: Data<ContaminationForm> = {
                    data: {
                        time: ctn.time,
                        confirmed: ctn.confirmed,
                        bacteria: ctn.bacteria,
                        mold: ctn.mold,
                        location: ctn.location,
                        notes: InitialToAllEntries(ctn.notes)
                    }, disabled: false
                }
                return out
            }),
        new: []
    }
}

// TODO: ensure used properly!!!
export function ContamsDisplay( // TODO: SET THIS UP SIMILARLY TO HOW PicsDisplay IS SET UP
    {
        readonly, initial, updateParent
    }: {
        readonly: boolean,
        initial: Contamination[],
        updateParent: (c: SplitAllEntries<ContaminationForm, NewContaminationForm>) => void,
        headerLevel?: number,
    }
) {
    const initFor = (inp: Contamination[]): Data<ContaminationForm>[] => {
        return inp.map(v => {
            return {
                data: {
                    time: v.time,
                    confirmed: v.confirmed,
                    bacteria: v.bacteria,
                    mold: v.mold,
                    location: v.location,
                    notes: InitialNotesState(v.notes),
                }, disabled: false
            }
        })
    }
    const [existing, setExisting] = useState<Data<ContaminationForm>[]>(initFor(initial))
    const [created, setCreated] = useState<NewContaminationForm[]>([])
    useEffect(() => {
        setExisting(initFor(initial))
        setCreated([])
        // TODO: update parent or no?
    }, [initial])
    const update = () => {
        updateParent({
            existing: structuredClone(existing),
            new: structuredClone(created),
        })
    }
    const doUpdateExisting = (updated: Data<ContaminationForm>[]) => {
        setExisting(updated)
        update()
    }
    const doUpdateNew = (updated: NewContaminationForm[]) => {
        setCreated(updated)
        update()
    }
    const depth = useContext(DepthContext)
    return <div className={"depthContainer depth" + depth}>
        <div className={"areaHeader"}>{"Contaminations:"}</div>
        <ContamsRows initial={initial} readonly={readonly} updateParent={doUpdateExisting}/>
        <ContamsNewRows initial={initial} readonly={readonly} updateParent={doUpdateNew}/>
    </div>
}

export function InitialDatum<T>(inp: T): Data<T> {
    return {data: inp, disabled: false}
}

export function ContamForm(v: Contamination): ContaminationForm {
    return {
        time: v.time,
        confirmed: v.confirmed,
        bacteria: v.bacteria,
        mold: v.mold,
        location: v.location,
        notes: InitialNotesState(v.notes)
    }
}

export function ContamsRows({initial, updateParent, readonly}: {
    initial: Contamination[],
    updateParent: (u: Data<ContaminationForm>[]) => void,
    readonly: boolean,
}) {
    const InitialData = (inp: Contamination[]): Data<ContaminationForm>[] => {
        return inp.map(v => {
            return {
                data: ContamForm(v),
                disabled: false,
            }
        })
    }
    const [current, setCurrent] = useState<Data<ContaminationForm>[]>(InitialData(initial))
    useEffect(() => {
        setCurrent(InitialData(initial))
    }, [initial])
    const disableButtonCreator = (i: number) => {
        if (readonly) {
            return null
        }
        return <RemoveToggle disabled={current[i].disabled} keptTxt={"Delete"} removedTxt={"Don't delete"}
                             keptClass={"removeButtonSmall"} removedClass={"basicButtonSmall"} click={() => {
            let updated = structuredClone(current)
            updated[i].disabled = !updated[i].disabled
            setCurrent(updated)
            updateParent(updated)
        }}/>
    }
    const doUpdate = (updated: Data<ContaminationForm>[]) => {
        setCurrent(updated)
        updateParent(updated)
    }
    return <div className={"contamsRows"}>
        {initial.map((init, i) => {
            const ctm = current[i]
            const disabledClass = (ctm.disabled ? " disabled" : "")
            const disableBtn = disableButtonCreator(i)
            return <div key={i} className={"contentsOnly contamRow" + disabledClass}>
                <div className={"picLeft" + disabledClass}>
                    {(ctm.data.location !== undefined) &&
                        <img className={/* TODO: IMAGE AREA GROW/SHRINK ON CLICK */"picDisplay"}
                             src={ImageLocationFor(ctm.data.location)}
                             alt={"existing contamination image " + i}/>
                    }
                    {disableBtn}
                </div>
                <div className={"contamOverviewTable" + disabledClass}>
                    <DateArea when={ctm.data.time} readonly={true}/>
                    {readonly ? <>
                        <div>{"Confirmed?: "+(ctm.data.confirmed ? "Confirmed" : "Unconfirmed")}</div>
                        <div>{"Bacteria: " + (ctm.data.bacteria ? "yes" : "no")}</div>
                        <div>{"Mold: " + (ctm.data.mold ? "yes" : "no")}</div>
                    </> : <>
                        <TestAndValidate todos={["toggle for confirmed and handle on the serverside"]}>
                            <div>{ctm.data.confirmed ? "Confirmed" : "Unconfirmed"}</div>
                        </TestAndValidate>
                        {/*<div>*/}{/* TODO: THIS!*/}
                        {/*    <div className={"inline"}>{"Confirmed? "}</div>*/}
                        {/*    <input className={"inline"} type={'checkbox'} disabled={init.confirmed}*/}
                        {/*           checked={ctm.data.confirmed}*/}
                        {/*           onChange={e => {*/}
                        {/*               e.stopPropagation();*/}
                        {/*               let updated = structuredClone(current)*/}
                        {/*               updated[i].data.confirmed = !ctm.data.confirmed*/}
                        {/*               doUpdate(updated)*/}
                        {/*           }}/>*/}
                        {/*</div>*/}
                        <div>
                            <div className={"inline"}>{"Bacteria: "}</div>
                            <input className={"inline"} type={'checkbox'} disabled={init.bacteria}
                                   checked={ctm.data.bacteria}
                                   onChange={e => {
                                       e.stopPropagation();
                                       let updated = structuredClone(current)
                                       updated[i].data.bacteria = !ctm.data.bacteria
                                       updated[i].data.confirmed == !ctm.data.bacteria || ctm.data.mold
                                       doUpdate(updated)
                                   }}/>
                        </div>
                        <div>
                            <div className={"inline"}>{"Mold: "}</div>
                            <input className={"inline"} type={'checkbox'} disabled={initial[i].mold}
                                   checked={ctm.data.mold}
                                   onChange={e => {
                                       let updated = structuredClone(current)
                                       const had = current[i].data
                                       updated[i].data.mold = !had.mold
                                       updated[i].data.confirmed == !had.mold || had.bacteria
                                       doUpdate(updated)
                                   }}/>
                        </div>
                    </>}
                </div>
                <div className={"inline" + disabledClass}>
                    <NotesFormArea readonly={readonly} initial={initial[i].notes} updateParent={nts => {
                        let updated = structuredClone(current)
                        updated[i].data.notes = nts
                        doUpdate(updated)
                    }}/>
                </div>
            </div>
        })}
    </div>
}

export function ContamsNewRows({initial, updateParent, readonly}: {
    initial: Contamination[],
    updateParent: (u: NewContaminationForm[]) => void,
    readonly: boolean,
}) {
    if (readonly) {
        return null
    }
    const [current, setCurrent] = useState<Data<NewContaminationForm>[]>([])
    useEffect(() => {
        setCurrent([])
    }, [initial])
    const update = (updated: Data<NewContaminationForm>[]) => {
        setCurrent(updated)
        updateParent(updated
            .filter(v => {
                const isEnabled = !v.disabled
                const hasNonDateData = (v.data.mold || v.data.bacteria || v.data.confirmed || v.data.file || v.data.notes.length > 0)
                return isEnabled && hasNonDateData
            })
            .map(v => v.data)
        )
    }
    return <>
        <div className={"contamsRows"}>
            {current.map((ctm, i) => {
                if (ctm.disabled) {
                    return null
                }
                return <div key={i} className={"contentsOnly contamRow"}>
                    <div className={"picLeft"}>
                        <ImageSelector updateParent={(f) => {
                            let updated = structuredClone(current)
                            updated[i].data.file = f
                            update(updated)
                        }}/>
                        <button className={"removeButtonSmall"} onClick={() => { // TODO: ensure works
                            let updated = structuredClone(current)
                            updated[i].disabled = true
                            update(updated)
                        }}>{"REMOVE THIS CONTAM"}</button>
                    </div>

                    {/* TODO: BUTTON ON NEXT LINE STYLING IS NO BUENO */}

                    <div className={"contamOverviewTable"}>
                        <DateArea when={ctm.data.time} readonly={true}/>
                        <TestAndValidate todos={["make confirmed changeable"]}>
                            <div>{ctm.data.confirmed ? "Confirmed" : "Unconfirmed"}</div>
                        </TestAndValidate>
                        {readonly ?
                            <div>{"Bacteria: " + (ctm.data.bacteria ? "yes" : "no")}</div> :
                            <div>
                                <div className={"inline"}>{"Bacteria: "}</div>
                                <input className={"inline"} type={'checkbox'} checked={ctm.data.bacteria}
                                       onChange={e => {
                                           let updated = structuredClone(current)
                                           updated[i].data.bacteria = !ctm.data.bacteria
                                           // TODO: CHANGE HOW CONFIRMED WORKS!
                                           updated[i].data.confirmed = ctm.data.bacteria || ctm.data.mold
                                           update(updated)
                                       }}/>
                            </div>
                        }
                        {readonly ?
                            <div>{"Mold: " + (ctm.data.mold ? "yes" : "no")}</div> /*TODO: THE ENTIRE READONLY PART!*/ :
                            <div>
                                <div className={"inline"}>{"Mold: "}</div>
                                <input className={"inline"} type={'checkbox'} checked={ctm.data.mold}
                                       onChange={e => {
                                           let updated = structuredClone(current)
                                           updated[i].data.mold = !ctm.data.mold
                                           // TODO: CHANGE HOW CONFIRMED WORKS!
                                           updated[i].data.confirmed = ctm.data.mold || ctm.data.bacteria
                                           update(updated)
                                       }}/>
                            </div>
                        }
                    </div>
                    <div className={"inline"}>
                        <NotesFormArea readonly={false} initial={[]} updateParent={nts => {
                            let updated = structuredClone(current)
                            updated[i].data.notes = nts.new.map(n => {
                                return n.data
                            })
                            update(updated)
                        }}/>
                    </div>
                </div>
            })}
        </div>
        {!readonly && <div className={"centerH gapTop"}>
            <button className={"greenButton"} onClick={() => {
                let updated = [...structuredClone(current), {
                    data: {
                        time: Date.now(), confirmed: false, file: undefined,
                        bacteria: false, mold: false, notes: [],
                    },
                    disabled: false
                }]
                update(updated)
            }}>{"Add New Contamination"}</button>
        </div>}
    </>
}

// TODO: ensure used properly!!!
// export function ContamsDisplayOld( // TODO: SET THIS UP SIMILARLY TO HOW PicsDisplay IS SET UP
//     {
//         readonly, headerLevel, initial, current, updateParent
//     }: {
//         readonly?: boolean,
//         initial: Contamination[],
//         current: SplitAllEntries<ContaminationForm, NewContaminationForm>, // TODO: probably requires overhaul to not use current!
//         updateParent: (c: SplitAllEntries<ContaminationForm, NewContaminationForm>) => void,
//         headerLevel?: number,
//     }
// ) {
//     const [count, setCount] = useState(0)
//     useEffect(() => {
//         setCount(count+1);
//     }, [initial]);
//     const doUpdateParent = (out: SplitAllEntries<ContaminationForm, NewContaminationForm>)=>{
//         let toParent = {...out}
//         toParent.new = toParent.new.filter(v=>!v.disabled)
//         updateParent(toParent)
//     }
//     const newImageSelected = (index: number) => {
//         return (f: File | undefined) => {
//             let data = {...current}
//             data.new[index].data.file = f
//             doUpdateParent(data)
//         }
//     }
//     const addNewFields = (e: React.MouseEvent) => {
//         e.preventDefault()
//         let data = {...current}
//         data.new = [...current.new, {
//             data: {
//                 time: Date.now(),
//                 confirmed: false,
//                 bacteria: false,
//                 mold: false,
//                 notes: [],
//             },
//             disabled: false
//         }]
//         doUpdateParent(data)
//     }
//     const removeNewFields = (index: number) => {
//             let data = {...current}
//             data.new[index].disabled = true
//             doUpdateParent(data)
//     }
//     const disableExistingFields = (index: number) => {
//             let data = {...current}
//             data.existing[index].disabled = !data.existing[index].disabled
//             doUpdateParent(data)
//     }
//     const depth = useContext(DepthContext)
//     const disableButtonCreator = (i: number)=>{
//        if (readonly){
//            return <div></div>
//        }
//        return <RemoveToggle disabled={current.existing[i].disabled} keptTxt={"Delete"} removedTxt={"Don't delete"} keptClass={"removeButtonSmall"} removedClass={"basicButtonSmall"} click={()=>{disableExistingFields(i)}}/>
//     }
//     // TODO: OVERHAUL WITH EITHER GRID OR FLEXBOX?
//     return <div key={count} className={"depthContainer depth"+depth}>
//         <div className={"areaHeader"}>{"Contaminations:"}</div>
//         <div className={"contamsRows"}>
//             {current.existing.map((ctm, i) => {
//
//                 const disableButton= disableButtonCreator(i)
//                 return <div key={i} className={"contentsOnly contamRow" + (ctm.disabled ? " disabled" : "")}>
//                     <div className={"picLeft" + (ctm.disabled ? " disabled" : "")}>
//                         {(ctm.data.location !== undefined) && <img className={/* TODO: IMAGE AREA GROW/SHRINK ON CLICK */"picDisplay"} src={ImageLocationFor(ctm.data.location)}
//                              alt={"existing contamination image " + i}/>
//                         }
//                         {disableButton}
//                     </div>
//                     <div className={"contamOverviewTable" + (ctm.disabled ? " disabled" : "")}>
//                         <DateArea when={ctm.data.time} readonly={true}/>
//                         <TestAndValidate todos={["toggle for confirmed and handle on the serverside"]}>
//                         <div>{ctm.data.confirmed ? "Confirmed" : "Unconfirmed"}</div>
//                         </TestAndValidate>
//                         {readonly ?
//                             <div>{"Bacteria: "+(ctm.data.bacteria?"yes":"no")}</div>:
//                             <div>
//                                 <div className={"inline"}>{"Bacteria: "}</div>
//                                 <input className={"inline"} type={'checkbox'} disabled={initial[i].bacteria} checked={ctm.data.bacteria}
//                                        onChange={e => {
//                                            let data = {...current}
//                                            data.existing[i].data.bacteria = !ctm.data.bacteria
//                                            data.existing[i].data.confirmed == !ctm.data.bacteria || ctm.data.mold
//                                            doUpdateParent(data)
//                                        }}/>
//                             </div>
//                         }
//                         {readonly ?
//                             <div>{"Mold: "+(ctm.data.mold?"yes":"no")}</div> :
//                             <div>
//                                 <div className={"inline"}>{"Mold: "}</div>
//                                 <input className={"inline"} type={'checkbox'} disabled={initial[i].mold} checked={ctm.data.mold}
//                                        onChange={e => {
//                                            let data = {...current}
//                                            data.existing[i].data.mold = !data.existing[i].data.mold
//                                            data.existing[i].data.confirmed == !data.existing[i].data.mold || data.existing[i].data.bacteria
//                                            doUpdateParent(data)
//                                        }}/>
//                             </div>
//                         }
//                     </div>
//                     <div className={"inline" + (ctm.disabled ? " disabled" : "")}>
//                         <NotesFormArea readonly={readonly} initial={initial[i].notes} updateParent={nts=>{
//                             let out = structuredClone(current)
//                             out.existing[i].data.notes = nts
//                             doUpdateParent(out)
//                         }}/>
//                     </div>
//                 </div>
//             })}
//         </div>
//         {!readonly && <div className={"contamsRows"}>
//             {current.new.map((ctm, i) => {
//                 if(ctm.disabled){
//                     return null
//                 }
//                 return <div key={i} className={"contentsOnly contamRow"}>
//                     <div className={"picLeft"}>
//                         <ImageSelector updateParent={newImageSelected(i)}/>
//                         <button className={"removeButtonSmall"} onClick={()=>{removeNewFields(i)}}>{"REMOVE THIS CONTAM"}</button>
//                     </div>
//
//                     {/* TODO: BUTTON ON NEXT LINE STYLING IS NO BUENO */}
//
//                     <div className={"contamOverviewTable"}>
//                         <DateArea when={ctm.data.time} readonly={true}/>
//                         {/* TODO: MAKE CONFIRMED CHANGEABLE */}
//                         <div>{ctm.data.confirmed ? "Confirmed" : "Unconfirmed"}</div>
//                         {readonly ?
//                             <div>{"Bacteria: " + (ctm.data.bacteria ? "yes" : "no")}</div> :
//                             <div>
//                                 <div className={"inline"}>{"Bacteria: "}</div>
//                                 <input className={"inline"} type={'checkbox'} checked={ctm.data.bacteria}
//                                        onChange={e => {
//                                            let data = {...current}
//                                            data.new[i].data.bacteria = !ctm.data.bacteria
//                                            // TODO: CHANGE HOW CONFIRMED WORKS!
//                                            data.new[i].data.confirmed = ctm.data.bacteria || ctm.data.mold
//                                            doUpdateParent(data)
//                                        }}/>
//                             </div>
//                         }
//                         {readonly ?
//                             <div>{"Mold: " + (ctm.data.mold ? "yes" : "no")}</div> /*TODO: THE ENTIRE READONLY PART!*/ :
//                             <div>
//                                 <div className={"inline"}>{"Mold: "}</div>
//                                 <input className={"inline"} type={'checkbox'} checked={ctm.data.mold}
//                                        onChange={e => {
//                                            let data = {...current}
//                                            data.new[i].data.mold = !ctm.data.mold
//                                            // TODO: CHANGE HOW CONFIRMED WORKS!
//                                            data.new[i].data.confirmed = ctm.data.mold || ctm.data.bacteria
//                                            doUpdateParent(data)
//                                        }}/>
//                             </div>
//                         }
//                     </div>
//                     <div className={"inline"}>
//                         <NotesFormArea readonly={false} initial={[]} updateParent={nts=>{
//                             let out = structuredClone(current)
//                             out.new[i].data.notes = nts.new.map(n => {
//                                 return n.data
//                             })
//                             doUpdateParent(out)
//                         }}/>
//                     {/*    <NotesGrid readonly={false} current={{existing:[],new:current.new[i].data.notes.map(v=>{return {data:v,disabled:false}})}} updateParent={(ns) => {*/}
//                     {/*    let out = {...current} // TODO: notesFormArea*/}
//                     {/*    out.new[i].data.notes = ns.new.map(n => {*/}
//                     {/*        return n.data*/}
//                     {/*    })*/}
//                     {/*    updateParent && updateParent(out)*/}
//                     {/*}}/>*/}
//                     </div>
//                 </div>
//             })}
//         </div>}
//
//         {!readonly && <div className={"centerH gapTop"}>
//             <button className={"greenButton"} onClick={addNewFields}>{"Add New Contamination"}</button>
//         </div>}
//     </div>
// }


export const ExampleTime = new Date().getTime()
export const TestNote = () => {
    return {time: new Date().getTime(), note: "TEST_NOTE_TEXT_HERE"}
}
export const TestNotes: Note[] = [TestNote(), TestNote(), TestNote()]
export const aPic: PicWithNotesIncoming = {time: ExampleTime, notes: [...TestNotes], location: ExampleImageLocation}
export const ExamplePicsWithNotesIncoming: PicWithNotesIncoming[] = [aPic, aPic, aPic]
export const c: Contamination = {
    time: ExampleTime,
    location: "test.jpg",
    mold: true,
    bacteria: false,
    confirmed: true,
    notes: [...TestNotes]
}


export const ExampleContaminationUnconfirmed = {
    time: ExampleTime,
    confirmed: false,
    bacteria: true,
    mold: false,
    location: ExampleImageLocation,
    notes: TestNotes,
}

export const ExampleContaminationConfirmed = {
    time: ExampleTime,
    confirmed: true,
    bacteria: false,
    mold: true,
    location: ExampleImageLocation,
    notes: TestNotes,
}

export const ExampleContaminations: Contamination[] = [ExampleContaminationUnconfirmed, ExampleContaminationConfirmed]
