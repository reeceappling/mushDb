'use client'

import React, {JSX, useContext, useState} from "react";
import {IsValidNote, NewEntryNotes, Note, NotesFormArea} from "@/app/components/formSubcomponents/notes";
import {AddCreatedTriColFunction, AllEntries, OnViewCreatorQuadCol} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
import {LcRecipeData} from "@/app/components/lcRecipeServer";
import {
    IsValidLiquid,
    Liquid,
    LiquidEntriesGroupForNew, LiquidsAreaReadOnly
} from "@/app/components/formSubcomponents/liquids";
import {
    IsValidNutrient,
    Nutrient, NutrientsAreaReadOnly,
    NutrientsEntriesGroupForNew
} from "@/app/components/formSubcomponents/nutrients";
import {
    IsValidSugar,
    Sugar,
    SugarEntriesGroupForNew, SugarsAreaReadOnly
} from "@/app/components/formSubcomponents/sugars";
import {
    CreatedLinkFor,
    DisplayFormWrapper,
    DisplayInput, DoCreateRequest, DoUpdateRequest, ExistingDualSelector, FlexedArea, FlexedSinglesGroup,
    ListPageItems, ListPageTable, ListTableColumn, NewColumn,
    NewEntryFormWrapper,
    NewEntryInput, NumberToDateStr,
    OptionalArrayOfType,
    RequiredArrayOfType, RequiredKey,
} from "@/app/components/common";
import EntryLinkForId, {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {
    Additive,
    AdditiveEntriesGroupForNew, AdditivesAreaReadOnly,
    IsValidAdditive
} from "@/app/components/formSubcomponents/additives";
import {ErrorDisplay, NameArea, StandardArea} from "@/app/components/formSubcomponents/commonClient";
import {
    AclDisplay,
    MarshalAcl,
    TogglableAreaWithDepth,
    UnmarshalAcl
} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import TestAndValidate from "@/app/components/testing/untested";
import {NewLcForm} from "@/app/components/lcClient";
import {LcData} from "@/app/components/lcServer";
import {OnViewCreatorsTriColArea} from "@/app/components/formSubcomponents/ovc";
import {InitialNotesState} from "@/app/components/formSubcomponents/initialState";
import {allCookies, CookiesContext} from "@/app/components/formSubcomponents/cookiesContext/cookies";
import {NameModifiable} from "@/app/components/jarRecipeClient";
import {Grain} from "@/app/components/formSubcomponents/grains";

export function AssertLcRecipe(input: any): asserts input is LcRecipeData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }
    // required simple keys
    const requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['name', 'string'],
        ['standard', 'boolean'],
        ['lastUpdated', 'number'],
    ])
    for (const [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('Agar Recipe assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }

    // complex required keys
    const complexRequiredKeys = new Map<string, (v: any) => boolean>([
        //['acl', IsValidAcl]
    ])
    for (const [key, validator] of complexRequiredKeys) {
        if (!RequiredKey(key, input, validator)) {
            throw new Error('LcRecipe assertion failure: required key ' + key + ' was not valid');
        }
    }

    // complex required array keys
    const complexRequiredArrayKeys = new Map<string, (v: any) => boolean>([
        ['liquids', IsValidLiquid],
    ])
    for (const [key, validator] of complexRequiredArrayKeys) {
        if (!RequiredArrayOfType(key, input, validator)) {
            throw new Error('LcRecipe assertion failure: optional array key ' + key + ' was not valid. {' + JSON.stringify(input[key]) + '}{' + JSON.stringify(input) + '}');
        }
    }

    // complex optional array keys
    const complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['nutrients', IsValidNutrient],
        ['sugars', IsValidSugar],
        ['additives', IsValidAdditive],
        ['notes', IsValidNote],
    ])
    for (const [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('LcRecipe assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    // Unmarshal ACL
    if (!('acl' in input)) {
        throw 'ACL missing from input in asserter'
    }
    input.acl = UnmarshalAcl(input.acl)
    return
}

export default function LcRecipeDisplay(
    {
        id, readonly, data, headerLevel
    }: DisplayInput<LcRecipeData>) {
        const [initial, setInitial] = useState(data)

        const [recName, setRecName] = useState(initial.name)
        const [isStandard, setIsStandard] = useState(initial.standard)
        const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(initial.notes))
        const [acl, setAcl] = useState<ACL>(initial.acl)
        const [err, setErr] = useState<string | undefined>()
        const updateInitial = (updated: LcRecipeData) => {
            setInitial(updated)
            setRecName(updated.name)
            setIsStandard(updated.standard)
            setNotes(InitialNotesState(updated.notes))
            setAcl(updated.acl)
            setErr(undefined)
        }
        const cookies = useContext(CookiesContext)
        const lcRecipeSubmit = () => {
            const body: any = {
                name: recName,
                standard: isStandard,
                notes: notes,
                acl: MarshalAcl(acl),
            }
            DoUpdateRequest("lcRecipe",initial._id, body, AssertLcRecipe, allCookies(cookies))
                .then(v=>{
                    updateInitial(new LcRecipeData(v))
                })
                .catch(e=>{
                    setErr("failed to update initial: "+JSON.stringify(e))
                })
        }
        const ovcs: OnViewCreatorQuadCol[] = [
            {
                txt: "New LC from LcRecipe",
                newCreationArea: (onCreate: AddCreatedTriColFunction) => {
                    return <NewLcForm lcRecipeIn={initial} handlers={{
                        onCreate: (newItem: LcData) => {
                            return onCreate([{
                                typeText: "Liquid Culture Jar",
                                node: <CreatedLinkFor linkId={newItem._id} typ={"lc"}/>
                            }], false)
                        },
                        isTopLevel: false,
                    }}/>
                },
            }
        ]
        return (
            <DisplayFormWrapper entryType={"lcRecipe"}>
                <ErrorDisplay err={err}/>
                <ID props={{id:data._id, txt:"Liquid Culture Recipe", entryType:"lcRecipe"}}>
                        <NameModifiable initial={initial.name} readonly={readonly} updateParent={setRecName}/>
                    </ID>
                <OnViewCreatorsTriColArea OnViewCreators={ovcs} readonly={readonly}/>
                <FlexedArea>
                    <FlexedSinglesGroup>
                        <DateArea pre={"Last Updated: "} when={initial.lastUpdated} readonly={true}/>

                    </FlexedSinglesGroup>
                    <FlexedSinglesGroup>
                        <StandardArea isStandard={isStandard} setStandard={setIsStandard} readonly={readonly}
                                      headerLevel={headerLevel}/>

                    </FlexedSinglesGroup>
                </FlexedArea>

                <LiquidsAreaReadOnly values={initial.liquids}/>
                <NutrientsAreaReadOnly values={initial.nutrients}/>
                <SugarsAreaReadOnly values={initial.sugars}/>
                <AdditivesAreaReadOnly values={initial.additives}/>
                <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>

                <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
                    <AclDisplay initial={acl} readonly={readonly} updateParent={setAcl}/>
                </TogglableAreaWithDepth>
                {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e) => {
                    e.stopPropagation();
                    lcRecipeSubmit()
                }}>{"Update"}</button>}
            </DisplayFormWrapper>
        )
}

export function NewLcRecipeForm({handlers}: { handlers: NewEntryInput<LcRecipeData> }) {
    const [defaultLiquids, setDefaultLiquids] = useState<Liquid[]>([]);
    const [defaultNutrients, setDefaultNutrients] = useState<Nutrient[]>([]);
    const [defaultSugars, setDefaultSugars] = useState<Sugar[]>([]);
    const [defaultAdditives, setDefaultAdditives] = useState<Additive[]>([]);

    const [name, setName] = useState("")
    const [isStandard, setIsStandard] = useState(false)
    const [liquids, setLiquids] = useState<Liquid[]>([])
    const [nutrients, setNutrients] = useState<Nutrient[]>([])
    const [sugars, setSugars] = useState<Sugar[]>([])
    const [additives, setAdditives] = useState<Additive[]>([])
    const [notes, setNotes] = useState<Note[]>([])
    const [err, setErr] = useState<string | undefined>(undefined)
    const [templateSelectorOpen, setTemplateSelectorOpen] = useState<boolean>(false)
    const loadTemplate = (template: LcRecipeData) => {
        setDefaultLiquids(template.liquids)
        setDefaultNutrients(template.nutrients || [])
        setDefaultSugars(template.sugars || [])
        setDefaultAdditives(template.additives || [])

        setLiquids(template.liquids)
        setNutrients(template.nutrients || [])
        setSugars(template.sugars || [])
        setAdditives(template.additives || [])
    }

    const cookies = useContext(CookiesContext)
    const createEntry = (e: React.MouseEvent) => {
        e.preventDefault()
        if (name === "") {
            setErr("invalid name")
            return
        }
        const body: any = {
            name: name,
            standard: isStandard,
            liquids: liquids,
            nutrients: nutrients,
            sugars: sugars,
            additives: additives,
            notes: notes
        }
        DoCreateRequest("lcRecipe", body, AssertLcRecipe, allCookies(cookies))
            .then(v=>{
                handlers.onCreate ? handlers.onCreate(v) : console.log("no onCreate provided")
            })
            .catch(e=>{
                setErr(JSON.stringify(e))
            })
    }
    const templateRecipeSelector = () => {
        if (templateSelectorOpen) {
            return <LcRecipeSelector doSelect={(rec) => { // TODO: endpoint for getStandard
                if (rec === undefined) {
                    return
                }
                loadTemplate(rec)
                setTemplateSelectorOpen(false)
            }} allowCreate={false}/>
        } else {
            return <button className={"basicButton"} onClick={() => {
                setTemplateSelectorOpen(true)
            }}>{"Select a template recipe (optional)"}</button>
        }
    }
    return (
        <NewEntryFormWrapper entryType={"lcRecipe"}>
            <ErrorDisplay err={err}/>
            {templateRecipeSelector()}
            <NameArea classNames={"inlineChildren"} currentName={name} setName={setName} headerTxt={"Recipe name: "}
                      readonly={false}/>
            <StandardArea isStandard={isStandard} setStandard={setIsStandard} headerTxt={"Standard recipe? "}
                          readonly={false}/>
            <div>{"Liquids"}</div>
            <LiquidEntriesGroupForNew initial={defaultLiquids} updateParent={setLiquids}/>
            <div>{"Nutrients"}</div>
            <NutrientsEntriesGroupForNew initial={defaultNutrients} updateParent={setNutrients}/>
            <div>{"Sugars"}</div>
            <SugarEntriesGroupForNew initial={defaultSugars} updateParent={setSugars}/>
            <div>{"Additives"}</div>
            <AdditiveEntriesGroupForNew initial={defaultAdditives} updateParent={setAdditives}/>
            <NewEntryNotes setNotes={setNotes}/>
            {/* SUBMIT AREA */}
            <button className={"greenButton buttonFullWidth"} onClick={createEntry}>{"Create"}</button>
        </NewEntryFormWrapper>
    )
}

export function LcRecipeArea({lcRecipeId, headerLevel, offset}: {
    lcRecipeId?: string,
    headerLevel?: number,
    offset?: number
}) {
    let linkArea: JSX.Element | null = <div>{"unknown"}</div>
    if (lcRecipeId !== undefined) {
        const b58id = lcRecipeId
        linkArea = <EntryLinkForId props={{
            displayId: b58id,
            linkId: b58id,
            entryType: "lcRecipe",
            openInNewTab: false, // TODO: ok?
            }}/>
    }
    return <div className={"lcRecipeArea"}>
        <div>{"Liquid Culture Recipe ID: "}</div>
        {linkArea}
    </div>
}

export function LcRecipeListPageTable({data, onClick, withLink}: ListPageItems<LcRecipeData>) {
    let cols: ListTableColumn<LcRecipeData>[] = [
        NewColumn("ID", (v) => v._id, true),
        NewColumn("Name", (v) => v.name, true), // TODO: shortname?
        NewColumn("Liquids", (v) => {
            return <div>
                {v.liquids.map((l, i) => {
                    return <div key={l.name + i}>{l.name}</div>
                })}
            </div>
        }, true),
        NewColumn("Nutrients", (v) => {
            return <div>
                {v.nutrients && v.nutrients.map((v, i) => {
                    return <div key={v.nutrient + i}>{v.nutrient}</div>
                })}
            </div>
        }, true),
        NewColumn("Sugars", (v) => {
            return <div>
                {v.sugars && v.sugars.map((v, i) => {
                    return <div key={v.type + i}>{v.type}</div>
                })}
            </div>
        }, true),
        NewColumn("Additives", (v) => {
            return <div>
                {v.additives && v.additives.map((v, i) => {
                    return <div key={v.additive + i}>{v.additive}</div>
                })}
            </div>
        }, true),
        NewColumn("Last Updated", (v) => {
            return NumberToDateStr(v.lastUpdated)
        }) // TODO: fit?
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: LcRecipeData) => {
            return <EntryLinkWrapper props={{entry:v, openInNewTab: true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable className={"text-xs"} cols={cols} data={data} onClick={onClick} newClass={v=>{return new LcRecipeData(v)}}/>
}

export function LcRecipeSelectorTable({data, onClick}: ListPageItems<LcRecipeData>) {
    const cols: ListTableColumn<LcRecipeData>[] = [
        NewColumn("ID", (v) => v._id),
        NewColumn("Name", (v) => v.name), // TODO: shortname?
        NewColumn("Last Updated", (v) => {
            return NumberToDateStr(v.lastUpdated)
        }),
        NewColumn("Link", (v: LcRecipeData) => {
            return <EntryLinkWrapper props={{entry:v, openInNewTab: true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })
    ]
    return <ListPageTable className={"text-xs"} cols={cols} data={data} onClick={onClick} newClass={v=>{return new LcRecipeData(v)}}/>
}

export function LcRecipeSelector(
    {
        doSelect,
        allowCreate
    }: {
        doSelect: (val: LcRecipeData | undefined) => void,
        allowCreate?: boolean
    }) {
    const table = (items: LcRecipeData[]): JSX.Element => {
        return <LcRecipeSelectorTable data={items} onClick={doSelect}/>
    }

    return <ExistingDualSelector entryType={"lcRecipe"} entryTypes={"lcRecipes"} doSelect={doSelect}
                                 asserter={AssertLcRecipe}
                                 table={table}>
        {allowCreate && <NewLcRecipeForm handlers={{onCreate: doSelect, isTopLevel: false}}/>}
    </ExistingDualSelector>
}
