'use client'

import React, {JSX, useContext, useState} from "react";
import {IsValidNote, NewEntryNotes, Note, NotesFormArea} from "@/app/components/formSubcomponents/notes";
import {AddCreatedTriColFunction, AllEntries, OnViewCreatorQuadCol} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
import {
    CreatedLinkFor,
    DisplayFormWrapper,
    DisplayInput,
    DoCreateRequest,
    DoUpdateRequest,
    ExistingRecentSelector,
    FlexedArea,
    FlexedSinglesGroup,
    ListPageItems,
    ListPageTable,
    ListTableColumn,
    NewColumn,
    NewEntryFormWrapper,
    NewEntryInput,
    NumberToDateStr,
    OptionalArrayOfType,
    OptionalSimpleKey, RequiredKey
} from "@/app/components/common";
import {ErrorDisplay} from "@/app/components/formSubcomponents/commonClient";
import {GrainBatchData} from "./grainBatchServer";
import {JarRecipeArea, JarRecipeSelector} from "@/app/components/jarRecipeClient";
import {NumericalArea} from "@/app/components/formSubcomponents/numericInput";
import {JarRecipeData} from "@/app/components/jarRecipeServer";
import {NewJarForm} from "@/app/components/jarClient";
import {JarData} from "@/app/components/jarServer";
import {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {OnViewCreatorsTriColArea} from "@/app/components/formSubcomponents/ovc";
import {InitialNotesState} from "@/app/components/formSubcomponents/initialState";
import {allCookies, CookiesContext} from "@/app/components/formSubcomponents/cookiesContext/cookies";
import {
    AclDisplay, MarshalAcl, UnmarshalAcl,
    TogglableAreaWithDepth
} from "@/app/components/accessControlClient";
import { ACL } from "./accessControlServer";

// TODO: GRAIN BATCHES LIST IS NOT WORKING!
// TODO: ENSURE DISPLAY IS LOOKING GOOD
// TODO: also plugs not working at all
// TODO: list users also not working (all of this as of 5/7/26)

export function AssertGrainBatch(input: any): asserts input is GrainBatchData {
    // TODO: FIX THIS WHOLE FUNC
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }
    // required simple keys
    const requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['recipe', 'string'],
        ['creationDate', 'number'],
        ['lastUpdated', 'number'],
    ])
    for (const [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('Grain Batch assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }

    // optional simple keys
    const optionalSimpleKeys = new Map<string, string>([
        ['soakTimeHrs', 'number'],
        ['boilTimeMins', 'number'],
        ['dryTimeHours', 'number'],
    ])
    for (const [key, expType] of optionalSimpleKeys) {
        if (!OptionalSimpleKey(key, input, expType)) {
            throw new Error('Batch assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex required keys
    const complexRequiredKeys = new Map<string, (v: any) => boolean>([
        //['acl', IsValidAcl],
    ])
    for (const [key, validator] of complexRequiredKeys) {
        if (!RequiredKey(key, input, validator)) {
            throw new Error('GrainBatch assertion failure: required key ' + key + ' was not valid');
        }
    }

    // complex optional array keys
    const complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['notes', IsValidNote],
    ])
    for (const [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('Grain Batch assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    // Unmarshal ACL
    if (!('acl' in input)) {
        throw 'ACL missing from input in asserter'
    }
    input.acl = UnmarshalAcl(input.acl)
    return
}

export default function GrainBatchDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel
    }: DisplayInput<GrainBatchData>) {
    // TODO: DISPLAY IS NOT WORKING FOR LONG IDS! ex: https://mush.appli.ng/view/grainBatch/5A44RFrtnHxcLTdoF
        const [initial, setInitial] = useState(data)

        const [err, setErr] = useState<string | undefined>()

        // grain non-changeable (base grain)
        // name non-changeable
        const [soakTime, setSoakTime] = useState<number | undefined>(initial.soakTimeHrs)
        const [boilTime, setBoilTime] = useState<number | undefined>(initial.boilTimeMins)
        const [dryTime, setDryTime] = useState<number | undefined>(initial.dryTimeHours)
        const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(initial.notes))
        const [acl, setAcl] = useState<ACL>(initial.acl)
        const updateInitial = (updated: GrainBatchData) => {
            setInitial(updated)
            setSoakTime(updated.soakTimeHrs)
            setBoilTime(updated.boilTimeMins)
            setDryTime(updated.dryTimeHours)
            setNotes(InitialNotesState(updated.notes))
            setAcl(updated.acl)
            setErr(undefined)
        }
        const cookies = useContext(CookiesContext)
        const submit = () => {
            const body: any = {
                soakTimeHrs: soakTime,
                boilTimeMins: boilTime,
                dryTimeHrs: dryTime,
                notes: notes,
                acl: MarshalAcl(acl),
            }
            DoUpdateRequest("grainBatch", initial._id, body, AssertGrainBatch, allCookies(cookies))
                .then(v=>{
                    updateInitial(new GrainBatchData(v))
                })
                .catch(e=>{
                    setErr("failed to update initial: "+JSON.stringify(e))
                })
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
                txt: "Create Jars From Batch", // TODO: creates either PC-d or un-pc'd jars!
                // TODO: does this creation need a pcRun??? Can we do it before the run?
                // TODO: can items be added when creating a PC run?
                newCreationArea: (onCreate: AddCreatedTriColFunction) => {
                    return <NewJarForm grainBatchIn={initial} recipeIn={initial.recipe} handlers={{
                        onCreate: (newItem: JarData) => {
                            return onCreate([{
                                typeText: "Grain Jar",
                                node: <CreatedLinkFor linkId={newItem._id} typ={"jar"}/>
                            }], false)
                        },
                        isTopLevel: false,
                    }}/>
                },
            }
        ]
        return <DisplayFormWrapper entryType={"grainBatch"}>
            <ErrorDisplay err={err} headerLevel={headerLevel}/>
            <ID props={{id:data._id, txt:"Grain Batch", entryType:"grainBatch", linkPage:false, allowOpenMainPage:false}}/>
            <OnViewCreatorsTriColArea OnViewCreators={ovcs} readonly={readonly}/>
            <FlexedArea>
                <FlexedSinglesGroup>
                    <JarRecipeArea recipeId={data.recipe}/>
                    <DateArea pre={"Last Updated: "} when={initial.lastUpdated} readonly={true}/>
                </FlexedSinglesGroup>
                <FlexedSinglesGroup>
                    <div>
                        {"Soak time (hrs): "}{initial.soakTimeHrs ? <text>{initial.soakTimeHrs}</text> :
                        <NumericalArea value={soakTime!==undefined ? soakTime.toString() : undefined}
                                       onChange={handleFormChangeSoak} label="SoakTimeHrs" min={0} step={1}
                                       errorMessage={'invalid amount'}
                                       mode={"integer"} readonly={readonly}/>
                        }

                    </div>
                    <div>
                        {"Boil time (mins): "}{initial.boilTimeMins ? <text>{initial.boilTimeMins}</text> :
                        <NumericalArea value={boilTime ? boilTime.toString() : undefined}
                                       onChange={handleFormChangeBoil} label="BoilTimeMinutes" min={0} step={1}
                                       errorMessage={'invalid amount'}
                                       mode={"integer"} readonly={readonly}/>}
                    </div>
                    <div>
                        {"Dry time (hrs): "}{initial.dryTimeHours ? <text>{initial.dryTimeHours}</text> :
                        <NumericalArea value={dryTime ? dryTime.toString() : undefined}
                                       onChange={handleFormChangeDry} label="DryTimeHours" min={0} step={1}
                                       errorMessage={'invalid amount'}
                                       mode={"integer"} readonly={readonly}/>}
                    </div>
                </FlexedSinglesGroup>
            </FlexedArea>

            <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
            <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
                <AclDisplay initial={acl} readonly={readonly} updateParent={setAcl}/>
            </TogglableAreaWithDepth>
            {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e) => {
                e.stopPropagation();
                submit()
            }}>{"Update"}</button>}
        </DisplayFormWrapper>
}

export function NewGrainBatchForm({handlers, recipe}: {
    handlers: NewEntryInput<GrainBatchData>,
    recipe?: JarRecipeData
}) {
    const [jarRecipe, setJarRecipe] = useState(recipe)
    const [notes, setNotes] = useState<Note[]>([])
    const [acl, setAcl] = useState<ACL>({blanketPerm:true})
    const [err, setErr] = useState<string | undefined>()

    const cookies = useContext(CookiesContext)
    const newGrainBatchSubmit = () => {
        if (jarRecipe === undefined) {
            setErr("jarRecipe must exist")
            return
        }
        const body: any = {
            recipe: jarRecipe?._id,
            notes: notes,
            acl: MarshalAcl(acl),
        }
        DoCreateRequest("grainBatch", body, AssertGrainBatch, allCookies(cookies))
            .then(v=>{
                handlers.onCreate ? handlers.onCreate(v) : console.log("no onCreate provided")
            })
            .catch(e=>{
                setErr(JSON.stringify(e))
            })
    }
    return <NewEntryFormWrapper entryType={"grainBatch"}>
        <ErrorDisplay err={err}/>
        {recipe === undefined &&
            <JarRecipeSelector doSelect={setJarRecipe} allowCreate={handlers.isTopLevel}
                               creatorInPage={handlers.isTopLevel}/>}
        <NewEntryNotes setNotes={setNotes}/>
        <AclDisplay readonly={false} updateParent={setAcl} initial={acl} />
        <button className={"bottomButton greenButton"} onClick={(e) => {
            e.stopPropagation();
            newGrainBatchSubmit()
        }}>{"Update"}</button>
    </NewEntryFormWrapper>
}

export function GrainBatchListPageTable({data, onClick, withLink}: ListPageItems<GrainBatchData>) {
    let cols: ListTableColumn<GrainBatchData>[] = [
        NewColumn("ID", (v) => v._id, true),
        NewColumn("Created", (v) => {
            return NumberToDateStr(v.creationDate)
        }, true),
        NewColumn("Updated", (v) => {
            return NumberToDateStr(v.lastUpdated)
        }),
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: GrainBatchData) => {
            return <EntryLinkWrapper props={{entry:v, openInNewTab: true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable cols={cols} data={data} onClick={onClick} newClass={v=>{return new GrainBatchData(v)}}/>
}

export function GrainBatchSelectorTable({data, onClick}: ListPageItems<GrainBatchData>) {
    return <GrainBatchListPageTable data={data} onClick={onClick} withLink={true}/>
}

export function GrainBatchSelector(
    {
        doSelect,
        allowCreate
    }: {
        doSelect: (val: GrainBatchData | undefined) => void,
        allowCreate?: boolean
    }) {
    const table = (items: GrainBatchData[]): JSX.Element => {
        return <GrainBatchSelectorTable data={items} onClick={doSelect}/>
    }

    return <ExistingRecentSelector entryType={"grainBatch"} entryTypes={"grainBatches"} doSelect={doSelect}
                                   asserter={AssertGrainBatch}
                                   table={table}>
        {allowCreate && <NewGrainBatchForm handlers={{onCreate: doSelect, isTopLevel: false}}/>}
    </ExistingRecentSelector>
}