'use client'

import * as React from "react";
import {ChangeEvent, RefObject, SetStateAction, useEffect, useRef, useState} from "react";
import {AllEntries, Data, GroupProps, RevertableAreaProps} from "@/app/components/formSubcomponents/shared";
import DateArea, {NumberToDate} from "@/app/components/formSubcomponents/date";
import {RemoveToggle} from "@/app/components/formSubcomponents/commonClient";
import {dataFor, Subform} from "@/app/components/common";
import {InitialNotesState} from "@/app/components/formSubcomponents/initialState";

// TODO: USE THIS ONE LIKE EVERYWHERE FOR VIEWS!
// TODO: Imports and New should use NewNotesForm
export function NotesFormArea({
                                  readonly,
                                  initial,
                                  updateParent,
                                  removeHeader,
                                  allowLargeTextBox = true
                              }: { // TODO: add withDictaphone if possible? we only want the dictaphone in some edge cases
    readonly?: boolean,
    initial?: Note[],
    updateParent?: (entries: AllEntries<Note>) => void,
    removeHeader?: boolean,
    allowLargeTextBox?: boolean
}) {
    return <div className={"notesViewArea"}>
        {removeHeader || <div className={"areaHeader"}>{"Notes"}</div>}
        <NotesAreaViewSubcomponent initial={initial || []} readonly={readonly || false} updateParent={upd => {
            updateParent && updateParent(upd)
        }} allowLargeTextBox={allowLargeTextBox}/>
    </div>
}

export function NotesFormAreaPics({
                                      readonly,
                                      initial,
                                      updateParent,
                                      removeHeader,
                                      allowLargeTextBox = true
                                  }: { // TODO: add withDictaphone if possible? we only want the dictaphone in some edge cases
    readonly?: boolean,
    initial?: Note[],
    updateParent?: (entries: AllEntries<Note>) => void,
    removeHeader?: boolean,
    allowLargeTextBox?: boolean
}) {
    return <div>
        {removeHeader || <div className={"areaHeader"}>{"Notes"}</div>}
        <NotesAreaPicsViewSubcomponent initial={initial || []} readonly={readonly || false} updateParent={upd => {
            updateParent && updateParent(upd)
        }} allowLargeTextBox={allowLargeTextBox}/>
    </div>
}

export type Note = {
    time: number
    note: string
}

export function IsValidNote(input: any): boolean {
    return (
        typeof input === 'object' &&
        'time' in input && typeof input.time === 'number' &&
        'note' in input && typeof input.note === 'string'
    )
}

// function useForceUpdate() {
//     const [, setToggle] = useState(false)
//     return () => setToggle(toggle => !toggle)
// }

// TODO: NotesFormArea/NotesAreaSubcomponent instead
export default function NotesArea({ // TODO: CURRENTLY DOES NOT WORK PROPERLY WHEN SOME NOTES ARE DELETED, FIX!
                                      readonly,
                                      initial,
                                      updateParent,
                                  }: {
    readonly?: boolean,
    initial?: Note[], // TODO: ensure everywhere is using this properly
    updateParent?: (entries: AllEntries<Note>) => void,
}) {
    const [current, setCurrent] = useState<AllEntries<Note>>(InitialNotesState(initial));
    useEffect(() => {
        setCurrent(InitialNotesState(structuredClone(initial)))
    }, [initial]);
    const updateCurrent = (updated: AllEntries<Note>) => {
        setCurrent(updated)
        updateParent && updateParent(updated);
    }
    const existingArea = () => {
        if (!current || current.existing.length === 0) {
            return null
        }
        return <div className={"notes"}>
            {current.existing.map((n, i) => {
                return <div key={i} className={"" + (n.disabled ? " disabled" : "")}>
                    <SingleNote value={n} readonly={readonly} updateParent={nd => {
                        const out = structuredClone(current)
                        out.existing[i] = structuredClone(nd)
                        updateCurrent(out)
                    }}/>
                    {!readonly && <RemoveNoteButton disabled={current.existing[i].disabled} click={() => {
                        const out = structuredClone(current)
                        out.existing[i].disabled = !out.existing[i].disabled
                        updateCurrent(out)
                    }}/>}
                </div>
            })}</div>

    }
    const createNewNote = () => {
        return {disabled: false, data: {time: new Date().getTime(), note: "FIXME"}} // TODO: fixme
    }
    // TODO: 5/3/26 creating 3 new notes and deleting the second does not update the new notes visually properly
    const newArea = () => {
        if (readonly) {
            return null
        }
        return <div>
            {(current?.new || []).map((n, i) => {
                if (n.disabled) {
                    return null
                }
                return <div
                    key={i}> {/* TODO: should be able to rely on key for deletion because "deleted" new notes are still in-mem*/}
                    <SingleNote startEditing={true} updateParent={nd => {
                        const out = structuredClone(current)
                        out.new[i] = structuredClone(nd)
                        updateCurrent(out)
                    }}/>
                    <RemoveNewNoteButton click={() => {
                        const out = structuredClone(current)
                        out.new[i].disabled = true;
                        const toParent = structuredClone(out)
                        toParent.new = toParent.new.filter(item => !(item.disabled))
                        updateCurrent(toParent)
                    }}/>
                </div>
            })}
            <div>
                <button className={"basicButtonSmall"} onClick={(e) => { // TODO: button to create new
                    e.preventDefault()
                    if (!!current) {
                        const out = structuredClone(current)
                        out.new = [...current.new, createNewNote()]
                        updateCurrent(out)
                    } else {
                        updateCurrent({existing: [], new: [createNewNote()]})
                    }
                }}>{"Create Note NotesArea"}</button>
            </div>
        </div>
    }
    return <div>
        {existingArea()}
        {newArea()}
    </div>

}

function RemoveNoteButton({disabled, click}: { disabled: boolean, click: () => void }) {
    return <RemoveToggle disabled={disabled} click={click} keptTxt={"Delete Note"} removedTxt={"Don't Delete"}
                         keptClass={"removeButtonSmall"} removedClass={"basicButtonSmall"}/>
}

function RemoveNewNoteButton({click}: { click: () => void }) {
    return <RemoveNoteButton disabled={false} click={click}/>
}


export function NotesAreaViewSubcomponent({initial, updateParent, readonly, allowLargeTextBox}: {
    initial: Note[],
    readonly: boolean,
    updateParent: (entries: AllEntries<Note>) => void,
    allowLargeTextBox: boolean
}) {
    const [existing, setExisting] = useState<Data<Note>[]>(dataFor(initial))
    const [created, setCreated] = useState<Data<Note>[]>([])
    const [reloadCount, setReloadCount] = useState(0)
    const currentClone = () => {
        return {
            existing: structuredClone(existing),
            new: structuredClone(created)
        }
    }
    useEffect(() => {
        const newAll = InitialNotesState(initial)
        setExisting(newAll.existing) // Must be first!
        setCreated(newAll.new)
        setReloadCount(reloadCount + 1)
        deliverUpdatesToParent(newAll)
    }, initial)

    const deliverUpdatesToParent = (updated: AllEntries<Note>) => {
        updateParent && updateParent(structuredClone(updated))
    }
    const updateExisting = (updated: Data<Note>[]) => {
        setExisting(updated)
        const out = currentClone()
        out.existing = updated
        deliverUpdatesToParent(out)
    }
    const updateCreated = (updated: Data<Note>[]) => {
        setCreated(updated)
        const out = currentClone()
        out.new = updated
        deliverUpdatesToParent(out)
    }
    const existingArea = () => {
        if (initial.length <= 0) {
            return null
        }

        return <>
            {existing.map((n, i) => {
                return <>
                    {/*// return <div key={i} className={"existingNote" + (n.disabled ? " disabled" : "")}>*/}
                    <SingleNoteV3 key={i + "sn"} initial={initial[i]} readonly={readonly}
                                  updateParent={nd => {
                                      const updated = structuredClone(existing)
                                      updated[i] = structuredClone(nd)
                                      updateExisting(updated)
                                  }}/>
                    <div key={i + "rnb"}>{readonly || <RemoveNoteButton disabled={n.disabled} click={() => {// TODO: does this need to be in a div?
                        const updated = structuredClone(existing)
                        updated[i].disabled = !n.disabled
                        updateExisting(updated)
                    }}/>}</div>
                </>
            })}
        </>
    }
    return <div className={"notesAreaV2Grid"}>
        {existingArea()}
        <NewNotesSubArea count={reloadCount} readonly={readonly} updateParent={updateCreated}
                         allowLargeTextBox={allowLargeTextBox}/>
    </div>
}

export function NotesAreaPicsViewSubcomponent({initial, updateParent, readonly, allowLargeTextBox}: {
    initial: Note[],
    readonly: boolean,
    updateParent: (entries: AllEntries<Note>) => void,
    allowLargeTextBox: boolean
}) {
    const [existing, setExisting] = useState<Data<Note>[]>(dataFor(initial))
    const [created, setCreated] = useState<Data<Note>[]>([])
    const [reloadCount, setReloadCount] = useState(0)
    const currentClone = () => {
        return {
            existing: structuredClone(existing),
            new: structuredClone(created)
        }
    }
    useEffect(() => {
        const newAll = InitialNotesState(initial)
        setExisting(newAll.existing) // Must be first!
        setCreated(newAll.new)
        setReloadCount(reloadCount + 1)
        deliverUpdatesToParent(newAll)
    }, initial)

    const deliverUpdatesToParent = (updated: AllEntries<Note>) => {
        updateParent && updateParent(structuredClone(updated))
    }
    const updateExisting = (updated: Data<Note>[]) => {
        setExisting(updated)
        const out = currentClone()
        out.existing = updated
        deliverUpdatesToParent(out)
    }
    const updateCreated = (updated: Data<Note>[]) => {
        setCreated(updated)
        const out = currentClone()
        out.new = updated
        deliverUpdatesToParent(out)
    }
    const existingArea = () => {
        if (initial.length <= 0) {
            return null
        }
        return <>
            {existing.map((n, i) => {
                return <div key={i} className={"existingNote" + (n.disabled ? " disabled" : "")}>
                    <SingleNotePics initial={initial[i]} readonly={readonly} allowLargeBox={allowLargeTextBox}
                                    updateParent={nd => {
                                        const updated = structuredClone(existing)
                                        updated[i] = structuredClone(nd)
                                        updateExisting(updated)
                                    }}/>
                    {readonly || <RemoveNoteButton disabled={n.disabled} click={() => {// TODO: does this need to be in a div?
                        const updated = structuredClone(existing)
                        updated[i].disabled = !n.disabled
                        updateExisting(updated)
                    }}/>}
                </div>
            })}
        </>
    }
    return <div className={"notesAreaV2"}>
        {existingArea()}
        <NewNotesPicsSubArea count={reloadCount} readonly={readonly} updateParent={updateCreated}
                             allowLargeTextBox={allowLargeTextBox}/>
    </div>
}

export function NewNotesSubArea({count, readonly, updateParent, allowLargeTextBox}: {
    count: number,
    readonly: boolean,
    updateParent: (entries: Data<Note>[]) => void,
    allowLargeTextBox: boolean
}) {
    if (readonly) {
        return null
    }
    const [notes, setNotes] = useState<Data<Note>[]>([])
    useEffect(() => {
        setNotes([]);
    }, [count]);
    const propagateUpdate = (updated: Data<Note>[]) => {
        setNotes(updated)
        updateParent(structuredClone(updated).filter((item) => {
            return !item.disabled && item.data.note !== ""
        }))
    }
    const defaultNote = () => {
        return {data: {time: new Date().getTime(), note: ""}, disabled: false}
    }
    const createNewNote = (e: React.MouseEvent<HTMLButtonElement, MouseEvent>) => {
        //e.preventDefault()
        e.stopPropagation();
        // Do not update parent here. We don't want to propagate empty notes
        setNotes([...structuredClone(notes), defaultNote()])
    }
    return <>
        {notes.map((n, i) => {
            if (n.disabled) {
                return null
            }
            return <>
                {/*return <div key={i}>*/}
                <SingleNoteV3 key={i + "sn"} readonly={false} startEditing={true} isNewSubAreaInViewPage={true}
                              updateParent={nd => {
                                  const updated = structuredClone(notes)
                                  updated[i].data = structuredClone(nd.data)
                                  propagateUpdate(updated)
                              }}/>
                <RemoveNewNoteButton key={i + "rnnb"} click={() => { // TODO: does this need to be in a div?
                    const updated = structuredClone(notes)
                    updated[i].disabled = true
                    propagateUpdate(updated)
                }}/>
            </>
        })}
        <div className={"createNewNoteButton"}>
            <button className={"basicButtonSmall"} onClick={createNewNote}>{"Create New Note"}</button>
        </div>
    </>


}

export function NewNotesPicsSubArea({count, readonly, updateParent, allowLargeTextBox}: {
    count: number,
    readonly: boolean,
    updateParent: (entries: Data<Note>[]) => void,
    allowLargeTextBox: boolean
}) {
    if (readonly) {
        return null
    }
    const [notes, setNotes] = useState<Data<Note>[]>([])
    useEffect(() => {
        setNotes([]);
    }, [count]);
    const propagateUpdate = (updated: Data<Note>[]) => {
        setNotes(updated)
        updateParent(structuredClone(updated).filter((item) => {
            return !item.disabled && item.data.note !== ""
        }))
    }
    const defaultNote = () => {
        return {data: {time: new Date().getTime(), note: ""}, disabled: false}
    }
    const createNewNote = (e: React.MouseEvent<HTMLButtonElement, MouseEvent>) => {
        //e.preventDefault()
        e.stopPropagation();
        // Do not update parent here. We don't want to propagate empty notes
        setNotes([...structuredClone(notes), defaultNote()])
    }
    return <div>
        {notes.map((n, i) => {
            if (n.disabled) {
                return null
            }
            return <div key={i}>
                <SingleNotePics readonly={false} startEditing={true} allowLargeBox={allowLargeTextBox}
                                updateParent={nd => {
                                    const updated = structuredClone(notes)
                                    updated[i].data = structuredClone(nd.data)
                                    propagateUpdate(updated)
                                }}/>
                <RemoveNewNoteButton click={() => { // TODO: does this need to be in a div?
                    const updated = structuredClone(notes)
                    updated[i].disabled = true
                    propagateUpdate(updated)
                }}/>
            </div>
        })}
        <div>
            <button className={"basicButtonSmall"} onClick={createNewNote}>{"Create New Note"}</button>
        </div>
    </div>


}

// TODO: consider only using initial for notes, but parent stores current for updates!
// TODO: this one is working, but should we use NotesArea instead????
// TODO: NotesFormArea/NotesAreaSubcomponent instead
export function NotesAreaOld({ // TODO: CURRENTLY DOES NOT WORK PROPERLY WHEN SOME NOTES ARE DELETED, FIX!
                                 readonly,
                                 current,
                                 updateParent,
                             }: RevertableAreaProps<Note>) {
    const existingArea = () => {
        if (!current || current.existing.length === 0) {
            return null
        }
        return <div className={"notes"}>
            {current.existing.map((n, i) => {
                return <div key={i} className={"" + (n.disabled ? " disabled" : "")}>
                    <SingleNote value={n} readonly={readonly} updateParent={nd => {
                        const out = {...current}
                        out.existing = [...out.existing]
                        out.existing[i] = nd
                        updateParent && updateParent(out)
                    }}/>
                    {!readonly && <RemoveNoteButton disabled={current.existing[i].disabled} click={() => {
                        const out = structuredClone(current)
                        out.existing[i].disabled = !current.existing[i].disabled
                        updateParent && updateParent(out)
                    }}/>}
                </div>
            })}
        </div>

    }
    const newArea = () => {
        if (readonly) {
            return null
        }
        return <div>{/* TODO: NEW*/}
            {(current?.new || []).map((n, i) => {
                if (n.disabled) {
                    return null
                }
                return <div key={i}>
                    <SingleNote startEditing={true} updateParent={nd => {
                        const out = {...(current || {existing: [], new: []})}
                        out.new[i] = nd
                        updateParent && updateParent(out)
                    }}/>
                    <RemoveNewNoteButton click={() => {
                        const out = {...(current || {existing: [], new: []})}
                        out.new[i].disabled = true;
                        const toParent = {...out}
                        toParent.new = toParent.new.filter(item => !item.disabled)
                        updateParent && updateParent(toParent)
                    }}/>
                </div>
            })}
            <div>
                <button className={"basicButtonSmall"} onClick={(e) => {
                    e.preventDefault()
                    if (!!current) {
                        const out = {...current}
                        out.new = [...current.new, {disabled: false, data: {time: new Date().getTime(), note: "FIXME"}}]
                        updateParent && updateParent(out)
                    } else {
                        updateParent && updateParent({
                            existing: [],
                            new: [{disabled: false, data: {time: new Date().getTime(), note: "FIXME"}}]
                        })
                    }
                    // let out = {...(current || {existing:[], new: []})}
                    // out.new = [...out.new, {disabled: false, data: {time: new Date().getTime(), note: "FIXME"}}]
                    // updateParent && updateParent(out)
                }}>{"Create Note"}</button>
            </div>
        </div>
    }
    return <div>
        {existingArea()}
        {newArea()}
    </div>

}

// TODO: NotesFormArea/NotesAreaSubcomponent instead
export function NotesGrid({
                              readonly,
                              current,
                              updateParent,
                          }: RevertableAreaProps<Note>) {
    // TODO: control initial vs final so that updating initial reverts the whole thing to the new values
    const existingArea = () => {
        if (!current || current.existing.length === 0) {
            return null
        }
        return <div className={"notes"}>
            {current.existing.map((n, i) => {
                return <div key={i} className={"" + (n.disabled ? " disabled" : "")}>
                    <SingleNote value={n} readonly={readonly} updateParent={nd => {
                        const out = structuredClone(current)
                        out.existing = [...out.existing]
                        out.existing[i] = nd
                        updateParent && updateParent(out)
                    }}/>
                    {!readonly && <RemoveNoteButton disabled={current.existing[i].disabled} click={() => {
                        const out = structuredClone(current)
                        out.existing[i].disabled = !current.existing[i].disabled
                        updateParent && updateParent(out)
                    }}/>}
                </div>
            })}
        </div>

    }
    const newArea = () => {
        if (readonly) {
            return null
        }
        return <div>{/* TODO: NEW*/}
            {(current?.new || []).map((n, i) => {
                if (n.disabled) {
                    return null
                }
                return <div key={i}> {/* TODO: CANNOT RELY ON KEY FOR DELETION*/}
                    <SingleNote startEditing={true} updateParent={nd => {
                        const out = {...(current || {existing: [], new: []})}
                        out.new[i] = nd
                        updateParent && updateParent(out)
                    }}/>
                    <RemoveNewNoteButton click={() => {
                        const out = {...(current || {existing: [], new: []})}
                        out.new[i].disabled = true;
                        const toParent = {...out}
                        toParent.new = toParent.new.filter(item => !item.disabled)
                        updateParent && updateParent(toParent)
                    }}/>
                </div>
            })}
            <div>
                <button className={"basicButtonSmall"} onClick={(e) => {
                    e.preventDefault()
                    if (!!current) {
                        const out = {...current}
                        out.new = [...current.new, {disabled: false, data: {time: new Date().getTime(), note: "FIXME"}}]
                        updateParent && updateParent(out)
                    } else {
                        updateParent && updateParent({
                            existing: [],
                            new: [{disabled: false, data: {time: new Date().getTime(), note: "FIXME"}}]
                        })
                    }
                    // let out = {...(current || {existing:[], new: []})}
                    // out.new = [...out.new, {disabled: false, data: {time: new Date().getTime(), note: "FIXME"}}]
                    // updateParent && updateParent(out)
                }}>{"Create Note Old"}</button>
            </div>
        </div>
    }
    return <div>
        {existingArea()}
        {newArea()}
    </div>

}

export function NotesAreaMostRecentImage({ // TODO: CURRENTLY DOES NOT WORK PROPERLY WHEN SOME NOTES ARE DELETED, FIX!
                                             readonly,
                                             current,
                                             updateParent,
                                         }: RevertableAreaProps<Note>) {
    // TODO: control initial vs final so that updating initial reverts the whole thing to the new values
    const existingArea = () => {
        if (!current || current.existing.length === 0) {
            return null
        }
        return <div className={"notes"}>
            {current.existing.map((n, i) => {
                return <div key={i} className={"" + (n.disabled ? " disabled" : "")}>
                    <SingleNote value={n} readonly={readonly} updateParent={nd => {
                        const out = structuredClone(current)
                        out.existing = [...out.existing]
                        out.existing[i] = nd
                        updateParent && updateParent(out)
                    }}/>
                    {!readonly &&
                        <RemoveNoteButton disabled={current.existing[i].disabled} click={() => {
                            const out = structuredClone(current)
                            out.existing[i].disabled = !current.existing[i].disabled
                            updateParent && updateParent(out)
                        }}/>}
                </div>
            })}
        </div>

    }
    const newArea = () => {
        if (readonly) {
            return null
        }
        return <div>{/* TODO: NEW*/}
            {(current?.new || []).map((n, i) => {
                if (n.disabled) {
                    return null
                }
                return <div key={i}> {/* TODO: CANNOT RELY ON KEY FOR DELETION*/}
                    <SingleNote startEditing={true} updateParent={nd => {
                        const out = {...(structuredClone(current) || {existing: [], new: []})}
                        out.new[i] = nd
                        updateParent && updateParent(out)
                    }}/>
                    <RemoveNewNoteButton click={() => {
                        const out = {...(structuredClone(current) || {existing: [], new: []})}
                        out.new[i].disabled = true;
                        const toParent = structuredClone(out)
                        toParent.new = toParent.new.filter(item => !item.disabled)
                        updateParent && updateParent(toParent)
                    }}/>
                </div>
            })}
            <div>
                <button className={"basicButtonSmall"} onClick={(e) => {
                    e.preventDefault()
                    if (!!current) {
                        const out = {...current}
                        out.new = [...structuredClone(current.new), {
                            disabled: false,
                            data: {time: new Date().getTime(), note: "FIXME"}
                        }]
                        updateParent && updateParent(out)
                    } else {
                        updateParent && updateParent({
                            existing: [],
                            new: [{disabled: false, data: {time: new Date().getTime(), note: "FIXME"}}]
                        })
                    }
                    // let out = {...(current || {existing:[], new: []})}
                    // out.new = [...out.new, {disabled: false, data: {time: new Date().getTime(), note: "FIXME"}}]
                    // updateParent && updateParent(out)
                }}>{"Create Note"}</button>
            </div>
        </div>
    }
    return <div>
        {existingArea()}
        {newArea()}
    </div>

}

// // TODO: can this be made into a non-client component?
// export default function NotesAreaOLD({
//                                       readonly,
//                                       initialValues,
//                                       updateParent,
//                                       headerLevel,
//                                       headerLevelOffset
//                                   }: AreaProps<Note>) {
//     const existingArea = () => {
//         if (current.existing.length === 0) {
//             return null
//         }
//         return <div className={"spreadContainedV"}>
//             {current.existing.map((n, i) => {
//                 // TODO: FIX KEYS! NOT WORKING!
//                 return <div key={i} className={"noteWrapper inlineChildren"+(n.disabled?" disabled":"")}>
//                     <SingleNote initialValue={n} readonly={readonly} createInEditMode={false} updateParent={nd => {
//                         let out = {...current}
//                         out.existing = [...out.existing]
//                         out.existing[i] = nd
//                         updateParent && updateParent(out)
//                         setCurrent(out)
//                     }}/>
//                     {!readonly && <button className={"noteSpaced"} onClick={e => {
//                         e.preventDefault()
//                         let out = {...current}
//                         out.existing[i].disabled = !current.existing[i].disabled
//                         updateParent && updateParent(out)
//                         setCurrent(out)
//                     }}>{current.existing[i].disabled ? "Don't Delete" : "Delete Note"}</button>}
//                 </div>
//             })}
//         </div>
//     }
//     const newArea = () => {
//         if (readonly) {
//             return null
//         }
//         return <div className={"spreadContainedV"}>{/* TODO: NEW*/}
//             {current.new.map((n, i) => {
//                 if (n.disabled) {
//                     return null
//                 }
//                 return <div key={i}> {/* TODO: CANNOT RELY ON KEY FOR DELETION*/}
//                     <SingleNote createInEditMode={true} updateParent={nd => {
//                         let out = {...current}
//                         out.new[i] = nd
//                         updateParent && updateParent(out)
//                         setCurrent(out)
//                     }}/>
//                     <button className={"noteSpaced"} onClick={e => {
//                         e.preventDefault()
//                         let out = {...current}
//                         out.new[i].disabled = true;
//                         let toParent = {...out}
//                         toParent.new = toParent.new.filter(item => !item.disabled)
//                         updateParent && updateParent(toParent)
//                         setCurrent(out)
//                     }}>{"Delete Note"}</button>
//                 </div>
//             })}
//             <div className={"centerH"}>
//                 <button onClick={e => {
//                     e.preventDefault()
//                     let out = {...current}
//                     out.new = [...out.new, {disabled: false, data: {time: new Date().getTime(), note: ""}}]
//                     updateParent && updateParent(out)
//                     setCurrent(out)
//                 }}>{"Create Note"}</button>
//             </div>
//         </div>
//     }
//     return <div className={"spreadContainedV"}>
//         {existingArea()}
//         {newArea()}
//     </div>
//
// }

// TODO: DELETE ANYTHING USING THIS!
export function SingleNote( // TODO: notes need a pretty major overhaul
    {
        readonly,
        value,
        startEditing,
        updateParent,
    }: {
        readonly?: boolean
        value?: Data<Note>
        startEditing?: boolean
        updateParent?: (n: Data<Note>) => void
    }) {
    const [val, setVal] = useState(value || {data: {time: new Date().getTime(), note: ""}, disabled: false})
    const [editing, setEditing] = useState(startEditing || false)
    const handleChangeText = (event: ChangeEvent<HTMLInputElement>) => {
        const data = {...val};
        data.data.note = event.target.value
        updateParent && updateParent(data)
        setVal(data)
    }
    const handleChangeDate = (newDate: number) => {
        const data = {...val};
        data.data.time = newDate
        updateParent && updateParent(data)
        setVal(data)
    }
    const withNote = (a: React.ReactNode) => {
        return <div className={"note"}>
            <DateArea readonly={readonly || !editing} when={val.data.time} updateParent={handleChangeDate}/>
            {a}
            {(!readonly && !editing) && <button className={"basicButtonSmall"} onClick={() => {
                setEditing(!editing)
            }}>{"Edit Note"}</button>}
        </div>
    }
    if (readonly || !editing) {
        return withNote(<div>{val.data.note}</div>)
    }
    return withNote(
        <input name='txt' type="text" disabled={false}
               autoComplete="off" onChange={handleChangeText}
               value={val.data.note} placeholder={"new note"}
               className={"noteValue rounded-none border-2 border-gray-300 bg-input px-4 text-left text-sm font-normal text-gray-900 placeholder:text-gray-400 focus:bg-white focus:outline-none focus:outline-0 focus:[&:not(:invalid)]:border-blue-300"}
               onBlur={() => {
                   setEditing(false)
               }}
        />)
}

const useAutosizeTextArea = (textAreaRef: React.RefObject<HTMLTextAreaElement | null>, value: string) => {
    useEffect(() => {
        if (textAreaRef.current) {
            // 1. Reset height to get the true inner height
            textAreaRef.current.style.height = "auto";
            // 2. Set the height to scrollHeight
            textAreaRef.current.style.height = `${textAreaRef.current.scrollHeight}px`;

            // 3. Measure and set width based on text content
            const canvas = document.createElement('canvas');
            const ctx = canvas.getContext('2d');
            if (ctx) {
                // Get computed font style from the textarea
                const computedStyle = window.getComputedStyle(textAreaRef.current);
                ctx.font = `${computedStyle.fontSize} ${computedStyle.fontFamily}`;

                // Measure text width
                const textWidth = ctx.measureText(value || 'a').width;
                const paddingLeftRight = parseFloat(computedStyle.paddingLeft) + parseFloat(computedStyle.paddingRight);
                const borderLeftRight = parseFloat(computedStyle.borderLeftWidth) + parseFloat(computedStyle.borderRightWidth);

                // Set width with padding/border included
                const totalWidth = textWidth + paddingLeftRight + borderLeftRight;
                textAreaRef.current.style.width = `${totalWidth}px`;
            }
        }
    }, [textAreaRef, value]);
};


export function AutoGrowTextarea({initial, className, updateParent, onBlur, ref}: { // TODO: set min/max widths! Set background colors!
    initial: string,
    updateParent: (s: string) => void,
    className?: string
    onBlur?: () => void
    ref: RefObject<HTMLTextAreaElement | null>
}) {
    const [value, setValue] = useState(initial);
    //const textAreaRef: RefObject<HTMLTextAreaElement | null> = useRef<HTMLTextAreaElement | null>(null);
    useEffect(() => {
        setValue(initial)
    }, [initial]);


    useAutosizeTextArea(ref, value);

    return (
        <textarea
            ref={ref}
            onChange={e => {
                const upd = e.target.value
                setValue(upd);
                updateParent(upd)
            }}
            value={value}
            rows={1}
            onBlur={onBlur}
            placeholder="Type your note here..." // TODO: FIX!
            className={className || "noteValue w-fit rounded-none border-2 border-gray-300 bg-input px-4 text-left text-sm font-normal text-gray-900 placeholder:text-gray-400 focus:bg-white focus:outline-none focus:outline-0 focus:[&:not(:invalid)]:border-blue-300"}
            // className={"noteValue w-fit rounded-none border-2 border-gray-300 bg-input px-4 text-left text-sm font-normal text-gray-900 placeholder:text-gray-400 focus:bg-white focus:outline-none focus:outline-0 focus:[&:not(:invalid)]:border-blue-300 "+className}
            style={{
                display: "block", // TODO: del if needed
                resize: "none",
                overflow: "hidden",
                transition: "height 0.2s ease", // Smooth transition
                minHeight: "40px",
                maxWidth: "100%",
                minWidth: 0, // TODO: maybe del
                width: "fit-content",//"100%", // TODO: maybe revert
                boxSizing: "border-box",
            }}
        />
    );
}

export function SingleNoteV2(
    {
        initial,
        readonly,
        startEditing,
        updateParent,
        allowLargeBox = false,
    }: {
        initial?: Note
        readonly: boolean
        startEditing?: boolean
        updateParent?: (n: Data<Note>) => void
        allowLargeBox: boolean
    }) {
    const defaultInitialNote = () => {
        return {time: new Date().getTime(), note: ""}
    }
    const [val, setVal] = useState<Data<Note>>({data: initial || defaultInitialNote(), disabled: false})
    const [started, setStarted] = useState(false)
    const [editing, setEditing] = useState(startEditing ?? false)
    const maxSize = 12
    const [small, setSmall] = useState(val.data.note.length < maxSize)
    const smRef = useRef<HTMLInputElement | null>(null);
    const lgRef = useRef<HTMLTextAreaElement | null>(null);
    useEffect(() => {
        setVal({data: initial || defaultInitialNote(), disabled: false})
        // setEditing(!started && (startEditing || false)) // TODO: ???
        // if (!started){
        //     setStarted(true)
        // }
        if (!started) {
            setEditing(startEditing || false)
            setStarted(true)
        } else {
            setEditing(false)
        }
    }, [initial])
    // const smallNoteEditArea = <input ref={smRef} name='txt' type="text" disabled={false}
    //                                  autoComplete="off" value={val.data.note}
    //                                  placeholder={"new note"}
    //                                  className={small ? "noteValue w-fit rounded-none border-2 border-gray-300 bg-input px-4 text-left text-sm font-normal text-gray-900 placeholder:text-gray-400 focus:bg-white focus:outline-none focus:outline-0 focus:[&:not(:invalid)]:border-blue-300" : "hidden"}
    //                                  onBlur={() => {
    //                                      setEditing(false)
    //                                  }}
    //                                  onChange={(e) => { // TODO: is this properly propagating the last letter?
    //                                      e.stopPropagation();
    //                                      //e.preventDefault(); // TODO: ????
    //                                      const updated = structuredClone(val);
    //                                      updated.data.note = e.target.value
    //                                      handleChangeNote(updated)
    //                                  }}
    // />
    const largeNoteEditArea = <AutoGrowTextarea
        /*max-w-full*/
        className={"noteValue w-fit rounded-none border-2 border-gray-300 bg-input px-4 text-left text-sm font-normal text-gray-900 placeholder:text-gray-400 focus:bg-white focus:outline-none focus:outline-0 focus:[&:not(:invalid)]:border-blue-300"}
        /*className={small ? "hidden" : ""}*/ ref={lgRef}
        initial={val.data.note/* TODO: NEEDS THOROUGH DEBUGGING!*/}
        onBlur={() => {
            setEditing(false)
        }}
        updateParent={s => {
            const updated = structuredClone(val);
            updated.data.note = s
            handleChangeNote(updated)
        }}/>
    const handleChangeNote = (updated: Data<Note>) => {
        const shouldBeLarge = updated.data.note.length >= maxSize
        setVal(updated)
        updateParent && updateParent(updated)
        // if ((editing || (!started && startEditing)) && small == shouldBeLarge) { // TODO: ensure ok
        //     setSmall(!small)
        //     if (shouldBeLarge) {
        //         // TODO: ignores ok?
        //         // @ts-ignore
        //         lgRef.current.focus() // TODO: NOT WORKING!
        //     } else {
        //         // @ts-ignore
        //         smRef.current.focus() // TODO: NOT WORKING!
        //     }
        // }
    }

    return <div className={"note"}>
        <DateArea readonly={readonly || !editing} when={val.data.time} updateParent={(newDate) => {
            const updated = structuredClone(val);
            updated.data.time = newDate
            handleChangeNote(updated)
        }}/>
        {(!readonly && editing) ? (/*(!allowLargeBox || small) ? smallNoteEditArea :*/ largeNoteEditArea) : <>
            <div>{val.data.note}</div>
            <button className={"basicButtonSmall"} onClick={() => {
                setEditing(!editing)
            }}>
                {"Edit Note"}
            </button>
        </>}
    </div>
}

export function SingleNoteV3(
    {
        initial,
        readonly,
        startEditing,
        updateParent,
        isNewSubAreaInViewPage = false,
    }: {
        initial?: Note
        readonly: boolean
        startEditing?: boolean
        updateParent?: (n: Data<Note>) => void
        isNewSubAreaInViewPage?: boolean
    }) {
    const defaultInitialNote = () => {
        return {time: new Date().getTime(), note: ""}
    }
    const [val, setVal] = useState<Data<Note>>({data: initial || defaultInitialNote(), disabled: false})
    const [started, setStarted] = useState(false)
    const [editing, setEditing] = useState(startEditing ?? false)
    const maxSize = 12
    // TODO: const [small, setSmall] = useState(val.data.note.length < maxSize)
    // TODO: const smRef = useRef<HTMLInputElement | null>(null);
    const lgRef = useRef<HTMLTextAreaElement | null>(null);
    useEffect(() => {
        setVal({data: initial || defaultInitialNote(), disabled: false})
        if (!started) {
            setEditing(startEditing || false)
            setStarted(true)
        } else {
            setEditing(false)
        }
    }, [initial])
    // const smallNoteEditArea = <input ref={smRef} name='txt' type="text" disabled={false}
    //                                  autoComplete="off" value={val.data.note}
    //                                  placeholder={"new note"}
    //                                  className={small ? "noteValue w-fit rounded-none border-2 border-gray-300 bg-input px-4 text-left text-sm font-normal text-gray-900 placeholder:text-gray-400 focus:bg-white focus:outline-none focus:outline-0 focus:[&:not(:invalid)]:border-blue-300" : "hidden"}
    //                                  onBlur={() => {
    //                                      setEditing(false)
    //                                  }}
    //                                  onChange={(e) => { // TODO: is this properly propagating the last letter?
    //                                      e.stopPropagation();
    //                                      //e.preventDefault(); // TODO: ????
    //                                      const updated = structuredClone(val);
    //                                      updated.data.note = e.target.value
    //                                      handleChangeNote(updated)
    //                                  }}
    // />
    const largeNoteEditArea = <AutoGrowTextarea
        /*max-w-full*/

        className={(isNewSubAreaInViewPage ? "" : "") + "noteTextEditor noteValue w-fit rounded-none border-2 border-gray-300 bg-input px-4 text-left text-sm font-normal text-gray-900 placeholder:text-gray-400 focus:bg-white focus:outline-none focus:outline-0 focus:[&:not(:invalid)]:border-blue-300"}
        /*className={small ? "hidden" : ""}*/ ref={lgRef}
        initial={val.data.note/* TODO: NEEDS THOROUGH DEBUGGING!*/}
        onBlur={() => {
            setEditing(false)
        }}
        updateParent={s => {
            const updated = structuredClone(val);
            updated.data.note = s
            handleChangeNote(updated)
        }}/>
    const handleChangeNote = (updated: Data<Note>) => {
        const shouldBeLarge = updated.data.note.length >= maxSize
        setVal(updated)
        updateParent && updateParent(updated)
        // if ((editing || (!started && startEditing)) && small == shouldBeLarge) { // TODO: ensure ok
        //     setSmall(!small)
        //     if (shouldBeLarge) {
        //         // TODO: ignores ok?
        //         // @ts-ignore
        //         lgRef.current.focus() // TODO: NOT WORKING!
        //     } else {
        //         // @ts-ignore
        //         smRef.current.focus() // TODO: NOT WORKING!
        //     }
        // }
    }

    return <>
        <div><DateArea readonly={readonly || !editing} when={val.data.time} updateParent={(newDate) => {
            const updated = structuredClone(val);
            updated.data.time = newDate
            handleChangeNote(updated)
        }}/></div>
        {(!readonly && editing) ? <>
            <div className="noteTextCell">{largeNoteEditArea}</div>
            <div>{/*empty button area*/}</div>
        </> : <>
            <div className="noteTextCell noteTextDisplay">{val.data.note}</div>
            <button className={"basicButtonSmall"} onClick={() => {
                setEditing(!editing)
            }}>{"Edit Note"}</button>
        </>}
    </>
}

export function SingleNotePics(
    {
        initial,
        readonly,
        startEditing,
        updateParent,
        allowLargeBox = false,
    }: {
        initial?: Note
        readonly: boolean
        startEditing?: boolean
        updateParent?: (n: Data<Note>) => void
        allowLargeBox: boolean
    }) {
    const defaultInitialNote = () => {
        return {time: new Date().getTime(), note: ""}
    }
    const [val, setVal] = useState<Data<Note>>({data: initial || defaultInitialNote(), disabled: false})
    const [started, setStarted] = useState(false)
    const [editing, setEditing] = useState(startEditing ?? false)
    const maxSize = 12
    const [small, setSmall] = useState(val.data.note.length < maxSize)
    const smRef = useRef<HTMLInputElement | null>(null);
    const lgRef = useRef<HTMLTextAreaElement | null>(null);
    useEffect(() => {
        setVal({data: initial || defaultInitialNote(), disabled: false})
        // setEditing(!started && (startEditing || false)) // TODO: ???
        // if (!started){
        //     setStarted(true)
        // }
        if (!started) {
            setEditing(startEditing || false)
            setStarted(true)
        } else {
            setEditing(false)
        }
    }, [initial])
    // const smallNoteEditArea = <input ref={smRef} name='txt' type="text" disabled={false}
    //                                  autoComplete="off" value={val.data.note}
    //                                  placeholder={"new note"}
    //                                  className={small ? "noteValue w-fit rounded-none border-2 border-gray-300 bg-input px-4 text-left text-sm font-normal text-gray-900 placeholder:text-gray-400 focus:bg-white focus:outline-none focus:outline-0 focus:[&:not(:invalid)]:border-blue-300" : "hidden"}
    //                                  onBlur={() => {
    //                                      setEditing(false)
    //                                  }}
    //                                  onChange={(e) => { // TODO: is this properly propagating the last letter?
    //                                      e.stopPropagation();
    //                                      //e.preventDefault(); // TODO: ????
    //                                      const updated = structuredClone(val);
    //                                      updated.data.note = e.target.value
    //                                      handleChangeNote(updated)
    //                                  }}
    // />
    const largeNoteEditArea = <AutoGrowTextarea
        /*max-w-full*/
        className={"noteValue w-fit rounded-none border-2 border-gray-300 bg-input px-4 text-left text-sm font-normal text-gray-900 placeholder:text-gray-400 focus:bg-white focus:outline-none focus:outline-0 focus:[&:not(:invalid)]:border-blue-300"}
        /*className={small ? "hidden" : ""}*/ ref={lgRef}
        initial={val.data.note/* TODO: NEEDS THOROUGH DEBUGGING!*/}
        onBlur={() => {
            setEditing(false)
        }}
        updateParent={s => {
            const updated = structuredClone(val);
            updated.data.note = s
            handleChangeNote(updated)
        }}/>
    const handleChangeNote = (updated: Data<Note>) => {
        const shouldBeLarge = updated.data.note.length >= maxSize
        setVal(updated)
        updateParent && updateParent(updated)
        // if ((editing || (!started && startEditing)) && small == shouldBeLarge) { // TODO: ensure ok
        //     setSmall(!small)
        //     if (shouldBeLarge) {
        //         // TODO: ignores ok?
        //         // @ts-ignore
        //         lgRef.current.focus() // TODO: NOT WORKING!
        //     } else {
        //         // @ts-ignore
        //         smRef.current.focus() // TODO: NOT WORKING!
        //     }
        // }
    }

    return <div className={"note"}>
        <DateArea readonly={readonly || !editing} when={val.data.time} updateParent={(newDate) => {
            const updated = structuredClone(val);
            updated.data.time = newDate
            handleChangeNote(updated)
        }}/>
        {(!readonly && editing) ? (/*(!allowLargeBox || small) ? smallNoteEditArea :*/ largeNoteEditArea) : <>
            <div>{val.data.note}</div>
            <button className={"basicButtonSmall"} onClick={() => {
                setEditing(!editing)
            }}>
                {"Edit Note"}
            </button>
        </>}
    </div>
}

export function NewEntryNotes({setNotes, readonly = false}: {
    readonly?: boolean,
    setNotes?: (value: SetStateAction<Note[]>) => void
}) {
    //const [created, setCreated] = useState<Data<Note>[]>([])
    const [reloadCount, setReloadCount] = useState(0) // TODO: delete or use!
    // const currentClone = () => {
    //     return structuredClone(created)
    // }

    const deliverUpdatesToParent = (updated: Data<Note>[]) => {
        setNotes && setNotes(structuredClone(updated).filter(n => !n.disabled).map(n => n.data))
    }
    // const updateCreated = (updated: Data<Note>[]) => {
    //     //setCreated(updated)
    //     deliverUpdatesToParent(updated)
    // }
    return <Subform>
        <div className={"notesAreaV2Grid"}>
            <NewNotesSubArea count={reloadCount} readonly={readonly} updateParent={deliverUpdatesToParent}
                             allowLargeTextBox={true}/>
        </div>
    </Subform>
    // return <Subform>
    //
    //     <div>{"Notes"}</div>
    //     <NoteEntriesGroup preexisting={false} readonly={false} updateParent={v => {
    //         setNotes && setNotes(v.map(x => {
    //             return x.data
    //         }))
    //     }}/>
    // </Subform>
}

export function NoteEntriesGroup({
                                     initialEntries,
                                     preexisting,
                                     readonly,
                                     updateParent,
                                 }: GroupProps<Note>) {
    const [inputFields, setInputFields] = useState<Data<Note>[]>(initialEntries || [])
    const handleFormChangeText = (index: number, newVal: string) => {
        const data = [...(structuredClone(inputFields) || [])];
        data[index].data.note = newVal
        updateParent(data)
        setInputFields(data);
    }
    // TODO: NUMBER AS DATE AND VISE-VERSA
    // const handleFormChangeDate = (index: number, event: ChangeEvent<HTMLInputElement>) => {
    //     changeDate(index, Number(event.target.value))
    // }
    // const changeDateToNow = (index: number,) => {
    //     return (e: MouseEvent) => {
    //         e.preventDefault()
    //         changeDate(index, Date.now())
    //     }
    // }
    const changeDate = (index: number, newDate: number) => {
        const data = [...(inputFields || [])];
        data[index].data.time = newDate
        updateParent(data)
        setInputFields(data);
    }
    const addFields = () => {
        const data = [...(inputFields || []), {data: {time: Date.now(), note: ''}, disabled: false}] // TODO: FIX DEFAULT
        updateParent(data)
        setInputFields(data)
    }
    const removeFields = (index: number) => {
        return () => {
            const data = [...(inputFields || [])];
            data.splice(index, 1) // TODO: THIS WONT WORK PROPERLY WITH INDEX
            updateParent(data)
            setInputFields(data)
        }
    }
    const disableField = (index: number) => {
        return () => {
            //e.preventDefault()e: MouseEvent
            const data = [...(inputFields || [])]
            data[index].disabled = !data[index].disabled
            updateParent(data)
            setInputFields(data)
        }
    }
    const notesAreaClasses = () => {
        let out = preexisting ? "exists" : "new"
        if (readonly) {
            out += " readonly"
        } else {
            out += " editable"
        }
        return out
    }
    const noteClasses = (note: Data<Note>) => {
        let out = "note"
        if (note.disabled) {
            out += " disabled"
        } else {
            out += " enabled"
        }
        return out
    }
    return <div className={notesAreaClasses()}>{/* TODO: CLASS STYLINGS!!!! */}
        {((inputFields || [])).map((input, index) => {
            return (
                <div className={noteClasses(input)} key={index}>
                    {input.disabled ? "disabled" : null /* TODO: remove? */}
                    <DateArea pre={readonly ? undefined : "Date: "} when={input.data.time} readonly={readonly}
                              updateParent={(n) => {
                                  changeDate(index, n)
                              }}/>
                    {readonly ? <div className={"noteValue"}>{input.data.note}</div> :
                        <input name='txt' type="text" disabled={false}
                               autoComplete="off" onChange={e => {
                            handleFormChangeText(index, e.target.value)
                        }}
                               value={input.data.note} placeholder={"new note"}
                            /* TODO: NOTEVALUE OK ON NEXT LINE? OR AT ALTERNATE DIV ABOVE?*/
                               className={"noteValue rounded-none border-2 border-gray-300 bg-input px-4 text-left text-sm font-normal text-gray-900 placeholder:text-gray-400 focus:bg-white focus:outline-none focus:outline-0 focus:[&:not(:invalid)]:border-blue-300"}
                               onBlur={() => { // TODO: on focus elsewhere
                                   // TODO: do nothing?
                               }/* TODO: MULTILINE TEXT INPUT FOR VERY LARGE THINGS!*/}
                        />
                        // <TextBox readonly={false} label={"noteLabel"/* TODO: REPLACE TEXTBOX????*/}
                        //          value={input.data.note} fieldName={"noteLabel"}
                        //          updateTextHandler={(s) => handleFormChangeText(index, s)}/>
                    }
                    {readonly ? null :
                        <button onClick={() => {
                            preexisting ? disableField(index)() : removeFields(index)()
                        }}>
                            {preexisting ? (input.disabled ? "enable" : "disable") : "remove"}
                        </button>}
                </div>)
        })}
        {preexisting ? null :
            <button className={"basicButtonSmall"} onClick={addFields}>{"Add More.."}</button>}
    </div>
}

export function NotesAreaInline(
    {
        notes, headerLevel, header, offset
    }: {
        notes?: Note[],
        headerLevel?: number,
        header?: string,
        offset?: number,
    }) {
    return <div>
        {header && <div>{header}</div>}
        {(notes || []).map((n, i) => {
            return <div key={i}>
                {NumberToDate(new Date(n.time)) + " - " + n.note}
            </div>
        })}
    </div>
}