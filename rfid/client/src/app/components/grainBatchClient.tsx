'use client'

import React, {useEffect, useState} from "react";
import {IsValidNote, NewEntryNotes, Note, NotesAreaInline} from "@/app/components/formSubcomponents/notes";
import {
    AddCreatedTriColFunction,
    AllEntries,
    OnViewCreatorQuadCol
} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
import {
    AssertArrayResult,
    DisplayInput,
    HandleJsonResponse,
    InlineExpansionArea,
    InlineExpansionButton,
    InlineProps,
    InlineSubArea, ListPageItems,
    NewEntryInput,
    OptionalArrayOfType,
    OptionalSimpleKey
} from "@/app/components/common";
import {ErrorDisplay} from "@/app/components/formSubcomponents/commonClient";
import {BaseExternalUrl} from "@/app/components/Constants";
import {GrainBatchData} from "./grainBatchServer";
import {JarRecipeArea, JarRecipeSelector} from "@/app/components/jarRecipeClient";
import {NumericalArea} from "@/app/components/formSubcomponents/numericInput";
import TestAndValidate from "@/app/components/testing/untested";
import {InitialNotesState} from "@/app/components/formSubcomponents/contaminations";
import {SelectorProps} from "@/app/components/selector";
import {JarRecipeData} from "@/app/components/jarRecipeServer";
import {OnViewCreatorsTriColArea} from "@/app/components/pcRunClient";
import {
    FlexedArea,
    FlexedSinglesGroup, ListPageTable,
    ListTableColumn,
    NewColumn,
    NotesFormArea, NumberToDateStr
} from "@/app/components/agarBatchClient";
import {CreatedLinkFor} from "@/app/components/substrateRecipeClient";
import {NewJarForm} from "@/app/components/jarClient";
import {JarData} from "@/app/components/jarServer";
import {DisplayFormWrapper, NewEntryFormWrapper} from "@/app/components/lcRecipeClient";
import {InlineEntry} from "./agarRecipeClient";

// TODO: GRAIN BATCHES LIST IS NOT WORKING!
// TODO: ENSURE DISPLAY IS LOOKING GOOD

export function AssertGrainBatch(input: any): asserts input is GrainBatchData {
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

export default function GrainBatchDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel, cookies
    }: DisplayInput) {
    try {
        AssertGrainBatch(data)
        const [initial, setInitial] = useState(data)

        const [err, setErr] = useState<string | undefined>()
        // TODO: THIS WHOLE THING!
        // TODO: only allow setting unset values once

        // grain non-changeable (base grain)
        // name non-changeable
        const [soakTime, setSoakTime] = useState<number | undefined>(initial.soakTimeHrs)
        const [boilTime, setBoilTime] = useState<number | undefined>(initial.boilTimeMins)
        const [dryTime, setDryTime] = useState<number | undefined>(initial.dryTimeHours)
        const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(initial.notes))
        const updateInitial = (updated: GrainBatchData) => {
            setInitial(updated)
            setSoakTime(updated.soakTimeHrs)
            setBoilTime(updated.boilTimeMins)
            setDryTime(updated.dryTimeHours)
            setNotes(InitialNotesState(updated.notes))
        }
        ////const [cookies, setCookie, removeCookie] = useCookies(['SessionId']);
        const submit = () => {
            fetch(BaseExternalUrl + "/db/update/grainBatch/" + data._id, {
                method: "POST",
                headers: {
                    credentials: 'include',
                    'Content-type': "application/json"
                },
                body: JSON.stringify({
                    soakTimeHrs: soakTime,
                    boilTimeMins: boilTime,
                    dryTimeHours: dryTime,
                    notes: notes,
                })
            })
                .then(HandleJsonResponse)
                .then((updated) => {
                    AssertGrainBatch(updated)
                    updateInitial(updated)
                })
                .catch((error) => {
                    setErr(JSON.stringify(error))
                });
        }
        const handleFormChangeBoil = (val?: string) => {
            const n = Number(val)
            if (Number.isNaN(n)) {
                setErr("NaN input for boil time")
            } else {
                val && setBoilTime(n)
                setErr(undefined)
            }
        }
        const handleFormChangeDry = (val?: string) => {
            const n = Number(val)
            if (Number.isNaN(n)) {
                setErr("NaN input for dry time")
            } else {
                val && setDryTime(n)
                setErr(undefined)
            }
        }
        const handleFormChangeSoak = (val?: string) => {
            const n = Number(val)
            if (Number.isNaN(n)) {
                setErr("NaN input for soak time")
            } else {
                val && setSoakTime(n)
                setErr(undefined)
            }
        }
        const ovcs: OnViewCreatorQuadCol[] = [
            {
                txt: "Create Jars From Batch", newCreationArea: (onCreate: AddCreatedTriColFunction) => {
                    return <NewJarForm grainBatchIn={initial} recipeIn={initial.recipe} handlers={{
                        onCreate: (newItem: JarData) => {
                            return onCreate([{
                                typeText: "Grain Jar",
                                node: <CreatedLinkFor linkId={newItem._id} typ={"jar"}/>
                            }])
                        },
                        isTopLevel: false,
                    }}/>
                }
            }
        ]
        return <DisplayFormWrapper entryType={"grainBatch"}>
            <TestAndValidate todos={["TEST THIS WHOLE THING!"]}><ErrorDisplay err={err} headerLevel={headerLevel}/>
                <ID id={id} txt={"Grain Batch"} entryType={"grainBatch"}/>
                <FlexedArea>
                    <FlexedSinglesGroup>{/*TODO: ALL THESE GROUPS!*/}
                        <JarRecipeArea recipeId={data.recipe}/>
                        <DateArea pre={"Last Updated: "} when={initial.lastUpdated} readonly={true}/>
                        <div>
                            {"Soak time (hrs): "}
                            <NumericalArea value={soakTime ? soakTime.toString() : undefined}
                                           onChange={handleFormChangeSoak} label="SoakTimeHrs" min={0} step={1}
                                           errorMessage={'invalid amount'}
                                           mode={"integer"} readonly={readonly}/>
                        </div>
                        <div>
                            {"Boil time (mins): "}
                            <NumericalArea value={boilTime ? boilTime.toString() : undefined}
                                           onChange={handleFormChangeBoil} label="BoilTimeMinutes" min={0} step={1}
                                           errorMessage={'invalid amount'}
                                           mode={"integer"} readonly={readonly}/>
                        </div>
                        <div>
                            {"Dry time (hrs): "}
                            <NumericalArea value={dryTime ? dryTime.toString() : undefined}
                                           onChange={handleFormChangeDry} label="DryTimeHours" min={0} step={1}
                                           errorMessage={'invalid amount'}
                                           mode={"integer"} readonly={readonly}/>
                        </div>
                    </FlexedSinglesGroup>
                </FlexedArea>

                <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
                {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e)=>{
                    e.stopPropagation();
                    submit()
                }}>{"Update"}</button>}
                <OnViewCreatorsTriColArea OnViewCreators={ovcs} readonly={readonly}/>{/* TODO: where to put?*/}
            </TestAndValidate>
        </DisplayFormWrapper>
    } catch (err) {
        return <div>{"ERROR: Grain Batch data format incorrect: " + err}</div>
    }
}

export function NewGrainBatchForm({handlers, recipe}: {
    handlers: NewEntryInput<GrainBatchData>,
    recipe?: JarRecipeData
}) {
    const [jarRecipe, setJarRecipe] = useState(recipe)
    const [notes, setNotes] = useState<Note[]>([])
    const [err, setErr] = useState<string | undefined>()
    // TODO: handle isTopLevel
    const newGrainBatchSubmit = () => {
        if (jarRecipe === undefined) {
            setErr("jarRecipe must exist")
            return
        }
        fetch(BaseExternalUrl + "/create/grainBatch", {
            method: 'Post',
            body: JSON.stringify({
                recipe: jarRecipe?._id,
                notes: notes,
            }),
            headers: {
                credentials: 'include',
                'Content-type': "application/json"
            },
        })
            .then(HandleJsonResponse)
            .then((newEntry) => {
                AssertGrainBatch(newEntry)
                handlers.onCreate && handlers.onCreate(newEntry)
            })
            .catch((err) => {
                setErr(JSON.stringify(err))
            });
    }
    return <NewEntryFormWrapper entryType={"grainBatch"}>
        <ErrorDisplay err={err}/>
        {recipe === undefined &&
            <JarRecipeSelector doSelect={setJarRecipe} headerLevel={3} allowCreation={handlers.isTopLevel}
                               creatorInPage={handlers.isTopLevel/* TODO: isTopLevel*/}/>}
        <NewEntryNotes setNotes={setNotes}/>
        <button className={"bottomButton greenButton"} onClick={(e)=>{
            e.stopPropagation();
            newGrainBatchSubmit()
        }}>{"Update"}</button>
    </NewEntryFormWrapper>
}

export function GrainBatchInline({
                                     data,
                                     expandByDefault,
                                     onClick,
                                     showMainPageButton,
                                     idIsLink
                                 }: InlineProps<GrainBatchData>) { // TODO: DO THIS ENTIRELY!
    const [expanded, setExpanded] = useState(expandByDefault)
    const b58id = data._id
    return <InlineEntry onClick={onClick}>
        <InlineSubArea props={{}}>
            <ID id={b58id} txt={"Grain Batch"} entryType={"grainBatch"} allowOpenMainPage={showMainPageButton}
                linkPage={idIsLink}/>
            <div>{"ADD CREATION DATE"}</div>
            {/* TODO: creation date! */}
            <div>{"ADD TIMINGS"}</div>
            {/* TODO: timings! */}
        </InlineSubArea>
        <InlineExpansionArea props={{expanded: expanded}}>
            <NotesAreaInline notes={data.notes}/>
            <DateArea pre={"Last Updated: "} when={data.lastUpdated} readonly={true}/>
        </InlineExpansionArea>
        <InlineExpansionButton data-cy-id="InlineSubAreaButton" setExpanded={setExpanded} expanded={expanded}/>
    </InlineEntry>
}

// export const GrainBatchArea = ({batchId, headerLevel}: { batchId?: string, headerLevel?: number }) => { // TODO: USE THIS!
//     let linkArea: JSX.Element | null = <div>{"unknown"}</div>
//     if (batchId !== undefined) {
//         const b58id = batchId
//         linkArea = <EntryLink props={{displayedId: b58id, linkId: b58id, entryType: "grainBatch"}}>
//             <div>{b58id}</div>
//             ]
//         </EntryLink>
//     }
//     return <div>
//         <div>{"Grain Batch ID: "}</div>
//         {linkArea}
//     </div>
// }

export function GrainBatchSelector(
    {
        doSelect, allowCreation, headerLevel, creatorInPage
    }: SelectorProps<GrainBatchData>) {
    // TODO: do these need depth providers?
    // TODO: HANDLE ALLOWCREATION AND CREATORINPAGE
    const [recent, setRecent] = useState<GrainBatchData[]>([])
    const [err, setErr] = useState<string | undefined>()
    useEffect(() => {
        fetch(BaseExternalUrl + "/db/list/grainBatches", {
            method: 'GET',
            headers: {
                credentials: 'include',
                'Content-type': "application/json"
            },
        })
            .then(HandleJsonResponse)
            .then((resp) => {
                AssertArrayResult<GrainBatchData>(resp, AssertGrainBatch)
                setRecent(resp)
            })
            .catch((err) => {
                setErr(JSON.stringify(err))
            });
    }, []) // TODO: OK????? [] or nothing?
    return <div>
        <ErrorDisplay err={err} headerLevel={headerLevel}/>
        <div>{"Recent Recipes"}</div>
        {recent.map(item => {
            return <GrainBatchInline data={item} headerLevel={headerLevel} onClick={doSelect} key={item._id}/>
        })}
        {/* TODO: CREATOR, IF ALLOWED, with increased depth */
        }
    </div>
}

export function GrainBatchListPageTable({data, onClick}: ListPageItems<GrainBatchData>) {
    const cols: ListTableColumn<GrainBatchData>[] = [
        NewColumn("ID", (v)=>v._id),
        NewColumn("Created", (v)=>{
            return NumberToDateStr(v.creationDate)
        }),
        NewColumn("Updated", (v)=>{
            return NumberToDateStr(v.lastUpdated)
        }),
    ]
    // TODO: expansion for everything else????
    return <ListPageTable cols={cols} data={data} onClick={onClick}/>
}