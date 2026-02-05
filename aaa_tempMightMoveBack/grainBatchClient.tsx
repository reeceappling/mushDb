'use client'

import {JSX, useEffect, useState} from "react";
import NotesArea, {
    IsValidNote,
    Note,
    NoteEntriesGroup,
    NotesAreaInline
} from "@/app/components/formSubcomponents/notes";
import {AllEntries, Data} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
import {
    DisplayInput,
    HandleJsonResponse,
    InlineExpansionArea,
    InlineProps,
    InlineSubArea,
    NewEntryInput,
    OptionalArrayOfType,
    OptionalSimpleKey
} from "@/app/components/common";
import EntryLink from "@/app/components/formSubcomponents/entryLink";
import {ErrorDisplay, OpenMainPage} from "@/app/components/formSubcomponents/commonClient";
import {BaseExternalUrl} from "@/app/components/Constants";
import {SelectorProps} from "@/app/components/selector";
import {useCookies} from "react-cookie";
import {GrainBatch} from "./grainBatchServer";
import {JarRecipeArea} from "@/app/components/jarRecipeClient";
import {NumberField} from "@base-ui/react";
import {NumericalArea} from "@/app/components/formSubcomponents/numericInput";
import {NewEntryNotes} from "../rfid/client/src/app/components/formSubcomponents/notes";

function AssertGrainBatch(input: any): asserts input is GrainBatch {
    // TODO: FIX THIS WHOLE FUNC
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }
    // required simple keys
    let requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['recipe', 'string'],
        ['creationDate', 'string'],
        ['lastUpdated', 'number'],
    ])
    for (let [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('Grain Batch assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }

    // optional simple keys
    let optionalSimpleKeys = new Map<string, string>([
        ['soakTimeHrs', 'number'],
        ['boilTimeMins', 'number'],
        ['dryTimeHours', 'number'],
    ])
    for (let [key, expType] of optionalSimpleKeys) {
        if (!OptionalSimpleKey(key, input, expType)) {
            throw new Error('Batch assertion failure: optional key ' + key + ' was not valid');
        }
    }

    // complex optional array keys
    let complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['notes', IsValidNote],
    ])
    for (let [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('Grain Batch assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    return
}

function dataFormatFor<T>(init: T[] | undefined): Data<T>[] {
    return (init || []).map((v) => {
        return {data: v, disabled: false}
    })
}

export default function GrainBatchDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel
    }: DisplayInput) {
    try {
        AssertGrainBatch(data)
        const [initial, setInitial] = useState(data)
        const [current, setCurrent] = useState(data)
        const [err, setErr] = useState<string | undefined>()
        // TODO: THIS WHOLE THING!
        // TODO: only allow setting unset values once

        // grain non-changeable (base grain)
        // name non-changeable
        const [notes, setNotes] = useState<AllEntries<Note>>({existing: dataFormatFor(initial.notes), new: []})
        const [cookies, setCookie, removeCookie] = useCookies(['SessionId']);
        const submit = () => {
            fetch(BaseExternalUrl + "/db/update/grainBatch/" + data._id, {
                method: "POST",
                headers: {
                    credentials: 'include',
                    SessionId: cookies.SessionId,
                    'Content-type': "application/json"
                },
                body: JSON.stringify({
                    soakTimeHrs:current.soakTimeHrs,
                    boilTimeMins:current.boilTimeMins,
                    dryTimeHours:current.dryTimeHours,
                    notes: notes,
                })
            })
                .then(HandleJsonResponse)
                .then((newEntry) => {
                    try {
                        AssertGrainBatch(newEntry)
                        setInitial(newEntry)
                        // TODO: set non-initial, like notes
                        // TODO: handle errors
                    } catch (e) {
                        // TODO: something
                    }
                })
                .catch((error) => {
                    setErr(JSON.stringify(error))
                });
        }
        const handleFormChangeBoil = (val: number) => {
            let data = {...current}
            data.boilTimeMins = val
            setCurrent(data)
        }
        const handleFormChangeDry = (val: number) => {
            let data = {...current}
            data.dryTimeHours = val
            setCurrent(data)
        }
        const handleFormChangeSoak = (val: number) => {
            let data = {...current}
            data.soakTimeHrs = val
            setCurrent(data)
        }
        return <div>
            <ErrorDisplay err={err} headerLevel={headerLevel}/>
            <ID id={id} headerLevel={headerLevel}/><OpenMainPage isTopLevel={isTopLevel} type={"grainBatch"}
                                                                         linkId={id}/>
            <JarRecipeArea recipeId={data.recipe}/>
            {"Soak time (hrs): "}<NumericalArea value={current.soakTimeHrs ? current.soakTimeHrs.toString() : undefined} onChange={(val?:string)=>{val && handleFormChangeSoak(Number(val))}} label="SoakTimeHrs" min={0} step={1} errorMessage={'invalid amount'} mode={"integer"} readonly={readonly} />
            {"Boil time (mins): "}<NumericalArea value={current.boilTimeMins ? current.boilTimeMins.toString() : undefined} onChange={(val?:string)=>{val && handleFormChangeBoil(Number(val))}} label="BoilTimeMinutes" min={0} step={1} errorMessage={'invalid amount'} mode={"integer"} readonly={readonly} />
            {"Dry time (hrs): "}<NumericalArea value={current.dryTimeHours ? current.dryTimeHours.toString() : undefined} onChange={(val?:string)=>{val && handleFormChangeDry(Number(val))}} label="DryTimeHours" min={0} step={1} errorMessage={'invalid amount'} mode={"integer"} readonly={readonly} />

            <NotesArea readonly={readonly} initialValues={dataFormatFor(initial.notes)} updateParent={setNotes}/>
            <DateArea pre={"Last Updated: "} when={initial.lastUpdated} readonly={true}/>
            {readonly ? null : <input type="submit" value="Update" onSubmit={submit}/>}
        </div>
    } catch (err) {
        return <div>{"ERROR: Grain Batch data format incorrect: " + err}</div>

    }
}

export function NewGrainBatchForm({handlers, recipe}: { handlers: NewEntryInput<GrainBatch>, recipe: string }) {
    const [notes, setNotes] = useState<Note[]>([])
    const [err, setErr] = useState<string | undefined>()
    const [cookies, setCookie, removeCookie] = useCookies(['SessionId']);
    const newGrainBatchSubmit = () => {
        fetch(BaseExternalUrl + "/create/grainBatch", {
            method: 'Post',
            body: JSON.stringify({
                recipe: recipe,
                notes: notes,
            }),
            headers: {
                credentials: 'include',
                SessionId: cookies.SessionId,
                'Content-type': "application/json"
                //Authorization: tokenFetch,
            },
        })
            .then(HandleJsonResponse)
            .then((newEntry) => {
                try {
                    AssertGrainBatch(newEntry)
                    handlers.onCreate && handlers.onCreate(newEntry)
                } catch (err) {
                    throw (err)
                }
            })
            .catch((err) => {
                setErr(JSON.stringify(err))
            });
    }
    return <div>
        <ErrorDisplay err={err}/>
        <NewEntryNotes setNotes={setNotes} />
        <input type="submit" value="Update" onSubmit={newGrainBatchSubmit}/>
    </div>
}

export function GrainBatchInline({data, expandByDefault, onClick, headerLevel}: InlineProps<GrainBatch>) { // TODO: DO THIS ENTIRELY!
    const [expanded, setExpanded] = useState(expandByDefault)
    const b58id = data._id
    return <div>
        <InlineSubArea props={{expanded: expanded, setExpanded: setExpanded}}>
            <ID id={b58id} headerLevel={headerLevel} onClick={() => {
                onClick && onClick(data)
            }}/>
            <div>{"ADD CREATION DATE"}</div>
            {/* TODO: creation date! */}
            <div>{"ADD TIMINGS"}</div>
            {/* TODO: timings! */}
        </InlineSubArea>
        <InlineExpansionArea props={{expanded: expanded, setExpanded: setExpanded}}>
            <NotesAreaInline notes={data.notes} headerLevel={headerLevel}/>
            <DateArea pre={"Last Updated: "} when={data.lastUpdated} readonly={true}/>
        </InlineExpansionArea>
    </div>
}

export const GrainBatchArea = ({batchId, headerLevel}: { batchId?: string, headerLevel?: number }) => { // TODO: USE THIS!
    let linkArea: JSX.Element | null = <div>{"unknown"}</div>
    if (batchId !== undefined) {
        const b58id = batchId
        linkArea = <EntryLink props={{displayedId: b58id, linkId: b58id, entryType: "grainBatch"}}>
            <div>{b58id}</div>
            ]
        </EntryLink>
    }
    return <div>
        <div>{"Grain Batch ID: "}</div>
        {linkArea}
    </div>
}

export function GrainBatchSelector(
    {doSelect, allowCreation, headerLevel, creatorInPage}: SelectorProps<GrainBatch> // TODO: ALL PROPS
) {
    const [recent, setRecent] = useState<GrainBatch[]>([])
    const [err, setErr] = useState<string | undefined>()
    const [cookies, setCookie, removeCookie] = useCookies(['SessionId']);
    useEffect(() => {
        fetch(BaseExternalUrl + "db/list/grainBatches", {
            method: 'GET',
            headers: {
                credentials: 'include',
                SessionId: cookies.SessionId,
                'Content-type': "application/json"
                //Authorization: tokenFetch,
            },
        })
            .then(HandleJsonResponse)
            .then((resp) => {
                let out = resp as GrainBatch[]
                setRecent(out)
            })
            .catch((err) => {
                setErr(JSON.stringify(err))
            });
    }, []) // TODO: OK????? [] or nothing?
    return <div>
        <ErrorDisplay err={err} headerLevel={headerLevel}/>
        <div>{/* Recent Area*/}
            <div>{"Recent Batches"}</div>
            {recent.map(item => {
                return <GrainBatchInline data={item} headerLevel={headerLevel} onClick={doSelect}/>
            })}
        </div>
    </div>
}