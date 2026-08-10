"use client"
// TODO: MAY WANT TO REMOVE USE CLIENT
// non-client even though it uses state?

import {AllEntries, Data, InitialToAllEntries, SplitAllEntries} from "@/app/components/formSubcomponents/shared";
import {IsValidNote, Note, NotesFormAreaPics} from "@/app/components/formSubcomponents/notes";
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

// add helper near top of file
function contaminationKey(items: Contamination[]): string {
    return items.map((c) =>
        [
            c.time,
            c.confirmed ? 1 : 0,
            c.bacteria ? 1 : 0,
            c.mold ? 1 : 0,
            c.location || "",
            (c.notes || []).map((n) => `${n.time}:${n.note}`).join("^"),
        ].join("|")
    ).join("||");
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
                const out: Data<ContaminationForm> = {
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

export function ContamsDisplay(
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
    const initKey = contaminationKey(initial);
    useEffect(() => {
        setExisting(initFor(structuredClone(initial)));
        setCreated([]);
    }, [initKey]);
    const update = (ex: Data<ContaminationForm>[], nw: NewContaminationForm[]) => {
        updateParent({
            existing: structuredClone(ex),
            new: structuredClone(nw),
        })
    }
    const doUpdateExisting = (updated: Data<ContaminationForm>[]) => {
        setExisting(updated)
        update(updated, created)
    }
    const doUpdateNew = (updated: NewContaminationForm[]) => {
        setCreated(updated)
        update(existing, updated)
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
    // TODO: create func for ContamRow!
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
    const initKey = contaminationKey(initial);
    useEffect(() => {
        setCurrent(InitialData(initial));
    }, [initKey]);
    const doUpdate = (updated: Data<ContaminationForm>[]) => {
        setCurrent(updated)
        updateParent(updated)
    }
    const disableButtonCreator = (i: number) => {
        if (readonly) {
            return null
        }
        return <RemoveToggle disabled={current[i].disabled} keptTxt={"Delete"} removedTxt={"Don't delete"}
                             keptClass={"removeButtonSmall"} removedClass={"basicButtonSmall"} click={() => {
            const updated = structuredClone(current)
            updated[i].disabled = !updated[i].disabled
            doUpdate(updated)
        }}/>
    }

    return <div className={"contamsRows"}>
        {initial.map((init, i) => {
            const ctm = current[i]
            const disabledClass = (ctm.disabled ? " disabled" : "")
            const disableBtn = disableButtonCreator(i)
            return <div key={i} className={"contentsOnly contamRow" + disabledClass}>
                <div className={"picLeft" + disabledClass}>
                    {(ctm.data.location !== undefined) &&
                        // <Image className={/* TODO: IMAGE AREA GROW/SHRINK ON CLICK */"picDisplay"}
                        //      src={ImageLocationFor(ctm.data.location)}
                        //      alt={"existing contamination image " + i}/>
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
                                       const updated = structuredClone(current)
                                       updated[i].data.bacteria = !ctm.data.bacteria
                                       updated[i].data.confirmed == !ctm.data.bacteria || ctm.data.mold
                                       doUpdate(updated)
                                       console.log("updating bacteria "+i+ "to "+!ctm.data.bacteria )
                                   }}/>
                        </div>
                        <div>
                            <div className={"inline"}>{"Mold: "}</div>
                            <input className={"inline"} type={'checkbox'} disabled={init.mold}
                                   checked={ctm.data.mold}
                                   onChange={e => {
                                       const updated = structuredClone(current)
                                       updated[i].data.mold = !ctm.data.mold
                                       updated[i].data.confirmed == !ctm.data.mold || ctm.data.bacteria
                                       doUpdate(updated)
                                   }}/>
                        </div>
                    </>}
                </div>
                <div className={"inline" + disabledClass}>
                    <NotesFormAreaPics readonly={readonly} initial={init.notes} allowLargeTextBox={false/* TODO: OK?*/}
                                   updateParent={nts => {
                        const updated = structuredClone(current)
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
    const initKey = contaminationKey(initial);
    useEffect(() => {
        setCurrent([]);
    }, [initKey]);
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
                            const updated = structuredClone(current)
                            updated[i].data.file = f
                            update(updated)
                        }}/>
                        <button className={"removeButtonSmall"} onClick={() => { // TODO: ensure works
                            const updated = structuredClone(current)
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
                                           const updated = structuredClone(current)
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
                                           const updated = structuredClone(current)
                                           updated[i].data.mold = !ctm.data.mold
                                           // TODO: CHANGE HOW CONFIRMED WORKS!
                                           updated[i].data.confirmed = ctm.data.mold || ctm.data.bacteria
                                           update(updated)
                                       }}/>
                            </div>
                        }
                    </div>
                    <div className={"inline"}>
                        <NotesFormAreaPics readonly={false} initial={[]} allowLargeTextBox={false/* TODO: OK?*/}
                            updateParent={nts => {
                                const updated = structuredClone(current)
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
                const updated = [...structuredClone(current), {
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
