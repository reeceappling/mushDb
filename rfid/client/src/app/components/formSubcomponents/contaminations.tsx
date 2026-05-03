// TODO: THESE ARE ALL NON-CLIENT!

import {AllEntries, Data, InitialToAllEntries, SplitAllEntries} from "@/app/components/formSubcomponents/shared";
import {NotesAreaOld, IsValidNote, Note, NotesGrid} from "@/app/components/formSubcomponents/notes";
import {
    ImageLocationFor, PicWithNotesIncoming
} from "@/app/components/formSubcomponents/picWithNotes";
import ImageSelector from "@/app/components/formSubcomponents/imageSelector";
import {OptionalArrayOfType, OptionalSimpleKey} from "@/app/components/common";
import DateArea from "@/app/components/formSubcomponents/date";
import {dataFor} from "@/app/components/agarRecipeClient";
import {useContext, useEffect, useState} from "react";
import {DepthContext} from "@/app/components/formSubcomponents/depthContext/depth";

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
    time: number // TODO: NEW! HANDLE EVERYWHERE!
    confirmed: boolean,
    bacteria: boolean,
    mold: boolean,
    location: string,
    notes: AllEntries<Note> // TODO: NEW! HANDLE EVERYWHERE!
}

export interface NewContaminationForm {
    time: number, // TODO: NEW! HANDLE EVERYWHERE!
    confirmed: boolean,
    bacteria: boolean,
    mold: boolean,
    file?: File,  // TODO: NEW! HANDLE EVERYWHERE!
    notes: Note[] // TODO: NEW! HANDLE EVERYWHERE!
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

export function InitialNotesState(existingNotes?: Note[]): AllEntries<Note> {
    return {existing: dataFor(existingNotes), new: []}
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
export function ContamsDisplay(
    {
        readonly, headerLevel, initial, current, updateParent
    }: {
        readonly?: boolean,
        initial: Contamination[],
        current: SplitAllEntries<ContaminationForm, NewContaminationForm>,
        updateParent: (c: SplitAllEntries<ContaminationForm, NewContaminationForm>) => void,
        headerLevel?: number,
    }
) {
    const [count, setCount] = useState(0)
    useEffect(() => {
        setCount(count+1);
    }, [initial]);
    const doUpdateParent = (out: SplitAllEntries<ContaminationForm, NewContaminationForm>)=>{
        let toParent = {...out}
        toParent.new = toParent.new.filter(v=>!v.disabled)
        updateParent(toParent)
    }
    const newImageSelected = (index: number) => {
        return (f: File | undefined) => {
            let data = {...current}
            data.new[index].data.file = f
            doUpdateParent(data)
        }
    }
    const addNewFields = (e: React.MouseEvent) => {
        e.preventDefault()
        let data = {...current}
        data.new = [...current.new, {
            data: {
                time: Date.now(),
                confirmed: false,
                bacteria: false,
                mold: false,
                notes: [],
            },
            disabled: false
        }]
        doUpdateParent(data)
    }
    const removeNewFields = (index: number) => {
            let data = {...current}
            data.new[index].disabled = true
            doUpdateParent(data)
    }
    const disableExistingFields = (index: number) => {
            let data = {...current}
            data.existing[index].disabled = !data.existing[index].disabled
            doUpdateParent(data)
    }
    const depth = useContext(DepthContext)
    // TODO: OVERHAUL WITH EITHER GRID OR FLEXBOX?
    return <div key={count} className={"depthContainer depth"+depth}>
        <div className={"areaHeader"}>{"Contaminations:"}</div>
        <div className={"contamsRows"}>
            {current.existing.map((ctm, i) => {
                const disableButton= (!readonly ? <button className={ctm.disabled?"removeButtonSmall":"basicButtonSmall"} onClick={()=>{disableExistingFields(i)}}>
                    {(ctm.disabled ? "Don't delete" : "Delete") + " contam"}
                </button>:<div></div>)
                return <div key={i} className={"contentsOnly contamRow" + (ctm.disabled ? " disabled" : "")}>
                    <div className={"picLeft" + (ctm.disabled ? " disabled" : "")}>
                        {(ctm.data.location !== undefined) ? <>
                            {/* TODO: IMAGE AREA GROW/SHRINK ON CLICK */}
                            <img className={"picDisplay"} src={ImageLocationFor(ctm.data.location)}
                             alt={"existing contamination image " + i}/>
                        {disableButton}</>:disableButton}
                    </div>
                    <div className={"contamOverviewTable" + (ctm.disabled ? " disabled" : "")}>
                        <DateArea when={ctm.data.time} readonly={true}/>
                        <div>{ctm.data.confirmed ? "Confirmed" : "Unconfirmed"}</div>
                        {readonly ?
                            <div>{"Bacteria: "+(ctm.data.bacteria?"yes":"no")}</div>:
                            <div>
                                <div className={"inline"}>{"Bacteria: "}</div>
                                <input className={"inline"} type={'checkbox'} disabled={initial[i].bacteria} checked={ctm.data.bacteria}
                                       onChange={e => {
                                           let data = {...current}
                                           data.existing[i].data.bacteria = !ctm.data.bacteria
                                           data.existing[i].data.confirmed == !ctm.data.bacteria || ctm.data.mold
                                           doUpdateParent(data)
                                       }}/>
                            </div>
                        }
                        {readonly ?
                            <div>{"Mold: "+(ctm.data.mold?"yes":"no")}</div> /*TODO: THE ENTIRE READONLY PART!*/ :
                            <div>
                                <div className={"inline"}>{"Mold: "}</div>
                                <input className={"inline"} type={'checkbox'} disabled={initial[i].mold} checked={ctm.data.mold}
                                       onChange={e => {
                                           let data = {...current}
                                           data.existing[i].data.mold = !data.existing[i].data.mold
                                           data.existing[i].data.confirmed == !data.existing[i].data.mold || data.existing[i].data.bacteria
                                           doUpdateParent(data)
                                       }}/>
                            </div>
                        }
                    </div>
                    <div className={"inline" + (ctm.disabled ? " disabled" : "")}>
                        {/* TODO: FIX NOTES SPACING. NotesFormArea?*/}
                        <NotesAreaOld readonly={readonly}
                                      current={current.existing[i].data.notes}
                                      updateParent={(nts) => { // TODO: do we even want this?
                                       let out = {...current}
                                       out.existing[i].data.notes = nts
                                       doUpdateParent(out)
                                   }}/>
                    </div>
                </div>
            })}
        </div>
        {!readonly && <div className={"contamsRows"}>
            {current.new.map((ctm, i) => {
                if(ctm.disabled){
                    return null
                }
                return <div key={i} className={"contentsOnly contamRow"}>
                    <div className={"picLeft"}>
                        <ImageSelector updateParent={newImageSelected(i)}/>
                        <button className={"removeButtonSmall"} onClick={()=>{removeNewFields(i)}}>{"REMOVE THIS CONTAM"}</button>
                    </div>

                    {/* TODO: BUTTON ON NEXT LINE STYLING IS NO BUENO */}

                    <div className={"contamOverviewTable"}>
                        <DateArea when={ctm.data.time} readonly={true}/>
                        {/* TODO: MAKE CONFIRMED CHANGEABLE */}
                        <div>{ctm.data.confirmed ? "Confirmed" : "Unconfirmed"}</div>
                        {readonly ?
                            <div>{"Bacteria: " + (ctm.data.bacteria ? "yes" : "no")}</div> :
                            <div>
                                <div className={"inline"}>{"Bacteria: "}</div>
                                <input className={"inline"} type={'checkbox'} checked={ctm.data.bacteria}
                                       onChange={e => {
                                           let data = {...current}
                                           data.new[i].data.bacteria = !ctm.data.bacteria
                                           // TODO: CHANGE HOW CONFIRMED WORKS!
                                           data.new[i].data.confirmed = ctm.data.bacteria || ctm.data.mold
                                           doUpdateParent(data)
                                       }}/>
                            </div>
                        }
                        {readonly ?
                            <div>{"Mold: " + (ctm.data.mold ? "yes" : "no")}</div> /*TODO: THE ENTIRE READONLY PART!*/ :
                            <div>
                                <div className={"inline"}>{"Mold: "}</div>
                                <input className={"inline"} type={'checkbox'} checked={ctm.data.mold}
                                       onChange={e => {
                                           let data = {...current}
                                           data.new[i].data.mold = !ctm.data.mold
                                           // TODO: CHANGE HOW CONFIRMED WORKS!
                                           data.new[i].data.confirmed = ctm.data.mold || ctm.data.bacteria
                                           doUpdateParent(data)
                                       }}/>
                            </div>
                        }
                    </div>
                    <div className={"inline"}>
                    <NotesGrid readonly={false} current={{existing:[],new:current.new[i].data.notes.map(v=>{return {data:v,disabled:false}})}} updateParent={(ns) => {
                        let out = {...current} // TODO: notesFormArea
                        out.new[i].data.notes = ns.new.map(n => {
                            return n.data
                        })
                        updateParent && updateParent(out)
                    }}/>
                    </div>
                </div>
            })}
        </div>}

        {!readonly && <div className={"centerH gapTop"}>
            <button className={"basicButton"} onClick={addNewFields}>{"Add New Contamination"}</button>
        </div>}
    </div>
}


export const ExampleTime = new Date().getTime()
export const TestNote = ()=>{
    return {time: new Date().getTime(), note:"TEST_NOTE_TEXT_HERE"}
}
export const TestNotes: Note[] = [TestNote(), TestNote(), TestNote()]
export const aPic: PicWithNotesIncoming = {time: ExampleTime, notes: [...TestNotes], location: ExampleImageLocation}
export const ExamplePicsWithNotesIncoming: PicWithNotesIncoming[] = [aPic,aPic,aPic]
export const c: Contamination = {time: ExampleTime, location: "test.jpg", mold:true, bacteria:false, confirmed:true, notes: [...TestNotes]}


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
