'use client'

import React, {JSX, useContext, useEffect, useState} from "react";
import {IsValidNote, NewEntryNotes, Note, NotesFormArea} from "@/app/components/formSubcomponents/notes";
import {AddCreatedTriColFunction, AllEntries, OnViewCreatorQuadCol} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
import {AgarRecipeData} from "@/app/components/agarRecipeServer";
import {
    IsValidLiquid,
    Liquid,
    LiquidEntriesGroupForNew,
    LiquidsAreaReadOnly
} from "@/app/components/formSubcomponents/liquids";
import {
    IsValidNutrient,
    Nutrient, NutrientsAreaReadOnly,
    NutrientsEntriesGroupForNew,
} from "@/app/components/formSubcomponents/nutrients";
import {
    IsValidSugar,
    Sugar,
    SugarEntriesGroupForNew,
    SugarsAreaReadOnly,
} from "@/app/components/formSubcomponents/sugars";
import {
    CreatedLinkFor,
    CreateNewEntryButton,
    dataFor,
    DisplayFormWrapper,
    DisplayInput,
    DoCreateRequest, DoGetRequest,
    DoUpdateRequest,
    ExistingDualSelector,
    FlexedArea,
    FlexedSinglesGroup,
    IsString,
    ListPageItems,
    ListPageTable,
    ListTableColumn,
    NewColumn,
    NewEntryFormWrapper,
    NewEntryInput,
    NumberToDateStr,
    OptionalArrayOfType,
    RequiredArrayOfType,
    RequiredKey, viewApiUrlFor,
    ViewInNewTabButton
} from "@/app/components/common";
import EntryLinkForId, {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {ErrorDisplay, InlineTitle, NameArea, StandardArea} from "@/app/components/formSubcomponents/commonClient";
import {
    Additive,
    AdditiveEntriesGroupForNew, AdditivesAreaReadOnly,
    IsValidAdditive
} from "@/app/components/formSubcomponents/additives";
import {
    Antibiotic,
    AntibioticEntriesGroupForNew,
    AntibioticsDisplay,
} from "@/app/components/formSubcomponents/antibiotic";
import TestAndValidate from "@/app/components/testing/untested";
import {AclDisplay, MarshalAcl, TogglableAreaWithDepth, UnmarshalAcl} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import {NewAgarBatchForm} from "@/app/components/agarBatchClient";
import {AgarBatchData} from "@/app/components/agarBatchServer";
import {InputNumber} from "@/app/components/formSubcomponents/numericInput";
import {OnViewCreatorsTriColArea} from "@/app/components/formSubcomponents/ovc";
import {InitialNotesState} from "@/app/components/formSubcomponents/initialState";
import {allCookies, CookiesContext} from "@/app/components/formSubcomponents/cookiesContext/cookies";
import {NameModifiable} from "@/app/components/jarRecipeClient";
import {BaseExternalUrl} from "@/app/components/Constants";

export function AssertAgarRecipe(input: any): asserts input is AgarRecipeData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }
    // required simple keys
    const requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['name', 'string'],
        ['agar', 'number'],
        ['standard', 'boolean'],
        ['lastUpdated', 'number'],
    ])
    for (const [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            console.error('Agar Recipe assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
            console.error(JSON.stringify(input));
            console.error(JSON.stringify(input[key]));
            throw new Error('Agar Recipe assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }

    // complex required keys
    const complexRequiredKeys = new Map<string, (v: any) => boolean>([
        //['acl', IsValidAcl]
    ])
    for (const [key, validator] of complexRequiredKeys) {
        if (!RequiredKey(key, input, validator)) {
            throw new Error('Agar Recipe assertion failure: required key ' + key + ' was not valid');
        }
    }

    // complex required array keys
    const complexRequiredArrayKeys = new Map<string, (v: any) => boolean>([
        ['liquids', IsValidLiquid],
    ])
    for (const [key, validator] of complexRequiredArrayKeys) {
        if (!RequiredArrayOfType(key, input, validator)) {
            console.error('AgarRecipe assertion failure: required array key ' + key + ' was not valid');
            console.error(JSON.stringify(input));
            console.error(JSON.stringify(input[key]));
            throw new Error('AgarRecipe assertion failure: required array key ' + key + ' was not valid');
        }
    }

    // complex optional array keys
    const complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['nutrients', IsValidNutrient],
        ['sugars', IsValidSugar],
        ['additives', IsValidAdditive],
        ['antibiotics', IsString],
        ['notes', IsValidNote],
    ])
    for (const [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            console.error('AgarRecipe assertion failure: optional array key ' + key + ' was not valid');
            console.error(JSON.stringify(input));
            console.error(JSON.stringify(input[key]));
            throw new Error('AgarRecipe assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    // Unmarshal ACL
    if (!('acl' in input)) {
        throw 'ACL missing from input in asserter'
    }
    input.acl = UnmarshalAcl(input.acl)
    return
}

export default function AgarRecipeDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel
    }:
    DisplayInput<AgarRecipeData>
) {
    const [initial, setInitial] = useState(data)
    // Required
    const [name, setName] = useState(data.name)
    // Optional
    const [isStandard, setIsStandard] = useState(data.standard)
    const [notes, setNotes] = useState<AllEntries<Note>>({existing: dataFor(data.notes || []), new: []})
    const [acl, setAcl] = useState<ACL>(data.acl)
    const [err, setErr] = useState<string | undefined>()
    const updateInitial = (updated: AgarRecipeData) => {
        setInitial(updated)
        setName(updated.name)
        setIsStandard(updated.standard)
        setNotes(InitialNotesState(updated.notes))
        setAcl(updated.acl)
        setErr(undefined)
    }
    const cookies = useContext(CookiesContext)

    const agarRecipeSubmit = () => {
        if (name === undefined || name === "") {
            setErr("Name field must not be empty")
            return
        }
        const body: any = {
            name: name,
            standard: isStandard,
            notes: notes,
            acl: MarshalAcl(acl),
        }
        DoUpdateRequest("agarRecipe", initial._id, body, AssertAgarRecipe, allCookies(cookies))
            .then(v => {
                updateInitial(new AgarRecipeData(v))
            })
            .catch(e => {
                setErr("failed to update initial: " + JSON.stringify(e))
            })
    }
    const ovcs: OnViewCreatorQuadCol[] = [
        {
            txt: "Create Batch From Recipe",
            newCreationArea: (onCreate: AddCreatedTriColFunction) => {
                return <NewAgarBatchForm agarRecipeIn={data} handlers={{
                    onCreate: (newItem: AgarBatchData) => {
                        return onCreate([{
                            typeText: "Agar Batch",
                            node: <CreatedLinkFor linkId={newItem._id} typ={"agarBatch"}/>
                        }], false) // TODO: true instead?
                    },
                    isTopLevel: false,
                }}/>
            },
        },
        {
            txt: "Create Plate (+batch)",
            newCreationArea: (onCreate: AddCreatedTriColFunction) => {
                return <TestAndValidate todos={["not implemented yet, should also create batch!", "Do MUCH later. Shortcut"]}>
                    <div>{"Not yet implemented!"}</div>
                </TestAndValidate>
            },
            needsTesting: true,
        },
        {
            txt: "Create Slant (+batch)",
            newCreationArea: (onCreate: AddCreatedTriColFunction) => {
                return <TestAndValidate todos={["not implemented yet, should also create batch!", "Do MUCH later. Shortcut"]}>
                    <div>{"Not yet implemented!"}</div>
                </TestAndValidate>
            },
            needsTesting: true,
        }
    ]
    return (
        <DisplayFormWrapper entryType={"agarRecipe"}>
            <ErrorDisplay err={err} headerLevel={headerLevel}/>
                <ID props={{id:data._id, txt:"Agar Recipe", entryType:"agarRecipe"}}>
                    <NameModifiable initial={initial.name} readonly={readonly} updateParent={setName}/>
                </ID>
            <OnViewCreatorsTriColArea OnViewCreators={ovcs} readonly={readonly}/>
            <FlexedArea>
                <FlexedSinglesGroup>
                    <DateArea pre={"Last Updated: "} when={initial.lastUpdated} readonly={true}/>
                </FlexedSinglesGroup>
                <FlexedSinglesGroup>
                    <StandardArea isStandard={isStandard} setStandard={setIsStandard} readonly={readonly}
                                  headerLevel={headerLevel}/>
                    <div className={"inlineChildren"}>
                        <div>{"Agar g/L: "}</div>
                        <div>{initial.agar}</div>
                    </div>
                </FlexedSinglesGroup>
            </FlexedArea>

            <LiquidsAreaReadOnly values={initial.liquids}/>
            <NutrientsAreaReadOnly values={initial.nutrients}/>
            <SugarsAreaReadOnly values={initial.sugars}/>
            <AdditivesAreaReadOnly values={initial.additives}/>
            <AntibioticsDisplay antibiotics={initial.antibiotics}/>
            <NotesFormArea readonly={readonly} initial={initial.notes}
                           updateParent={setNotes}/>{/* TODO: this is erroring when updating after creating a note, then erroring again when trying to click update with no changes because it says existing notes length is not the same!*/}
            <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
                <AclDisplay initial={acl} readonly={readonly}
                            updateParent={setAcl}/>{/*TODO: agarRecipe 1 is not properly loading the initial acl!*/}
            </TogglableAreaWithDepth>
            {readonly ? null :
                <button className={"bottomButton greenButton"} onClick={(e) => {
                    e.stopPropagation();
                    agarRecipeSubmit()
                }}>{"Update"}</button>}

        </DisplayFormWrapper>
    )
}

export function NewAgarRecipeForm({handlers}: { handlers: NewEntryInput<AgarRecipeData> }) {
    const [defaultLiquids, setDefaultLiquids] = useState<Liquid[]>([]);
    const [defaultNutrients, setDefaultNutrients] = useState<Nutrient[]>([]);
    const [defaultSugars, setDefaultSugars] = useState<Sugar[]>([]);
    const [defaultAdditives, setDefaultAdditives] = useState<Additive[]>([]);
    const [defaultAntibiotics, setDefaultAntibiotics] = useState<Antibiotic[]>([]);

    const [name, setName] = useState("")
    const [isStandard, setIsStandard] = useState(false)
    const [agar, setAgar] = useState(20)
    const [liquids, setLiquids] = useState<Liquid[]>([])
    const [nutrients, setNutrients] = useState<Nutrient[]>([])
    const [sugars, setSugars] = useState<Sugar[]>([])
    const [additives, setAdditives] = useState<Additive[]>([])
    const [antibiotics, setAntibiotics] = useState<Antibiotic[]>([])
    const [notes, setNotes] = useState<Note[]>([])
    const [err, setErr] = useState<string | undefined>()
    const [agarErr, setAgarErr] = useState<string | undefined>()
    const [templateSelectorOpen, setTemplateSelectorOpen] = useState<boolean>(false)
    const cookies = useContext(CookiesContext)
    const newAgarRecipeSubmit = () => {
        if (name === "") {
            setErr("name must not be empty")
            return
        }
        if (liquids.length === 0) {
            setErr("at least one liquid must exist")
            return
        }
        const body: any = {
            name: name,
            standard: isStandard,
            agar: agar,
            liquids: liquids,
            // Optional
            nutrients: nutrients.length !== 0 ? nutrients : undefined,
            sugars: sugars.length !== 0 ? sugars : undefined,
            additives: additives.length !== 0 ? additives : undefined,
            antibiotics: antibiotics.length !== 0 ? antibiotics : undefined,
            notes: notes.length !== 0 ? notes : undefined,
        }
        DoCreateRequest("agarRecipe", body, AssertAgarRecipe, allCookies(cookies))
            .then(v => {
                handlers.onCreate ? handlers.onCreate(v) : console.log("no onCreate provided")
            })
            .catch(e => {
                setErr(JSON.stringify(e))
            })
    }
    const templateRecipeSelector = () => {
        if (templateSelectorOpen) {
            return <AgarRecipeSelector doSelect={(rec) => {
                if (rec === undefined) {
                    return
                }
                setDefaultLiquids(rec.liquids)
                setDefaultNutrients(rec.nutrients || [])
                setDefaultSugars(rec.sugars || [])
                setDefaultAdditives(rec.additives || [])
                setDefaultAntibiotics(rec.antibiotics || [])

                setName(rec.name)
                setIsStandard(rec.standard)
                setAgar(rec.agar)
                setLiquids(rec.liquids)
                setNutrients(rec.nutrients || [])
                setSugars(rec.sugars || [])
                setAdditives(rec.additives || [])
                setAntibiotics(rec.antibiotics || [])
                // notes?
                setTemplateSelectorOpen(false)
            }}/>
        } else {
            return <button className={"basicButton"} onClick={() => {
                setTemplateSelectorOpen(true)
            }}>{"Select a template recipe (optional)"}</button>
        }
    }
    return (
        <NewEntryFormWrapper entryType={"agarRecipe"}>
            <ErrorDisplay err={err}/>
            <div>
                {templateRecipeSelector()}
            </div>
            <NameArea classNames={"inlineChildren"} titleClasses={"mr-2"} currentName={name || ""} setName={setName}
                      headerTxt={"Recipe Name: "} readonly={false}/>
            <StandardArea isStandard={isStandard} setStandard={setIsStandard} readonly={false}
                          headerTxt={"Standard Recipe? "}/>
            <div className={"inlineChildren my-4"}>
                <InlineTitle title={"Agar g/L: "} titleClasses={"mr-2"}>
                    <InputNumber readonly={false} value={"" + agar} min={0} max={100}
                                 mode={"integer"} onChange={(e) => {
                        const val = e || ""
                        try {
                            if (val === '' || /^\d*\.?\d*$/.test(val)) {
                                setAgar(parseFloat(val));
                            }
                        } catch {
                            setAgarErr("invalid agar amount")
                        }
                    }} placeholder={"10"} errorMessage={agarErr}/>
                    <div className={"ml-2"}>
                        {agarPer400mL(agar)}
                    </div>
                </InlineTitle>
            </div>
            {/* TODO: liquids and below as flexbox?*/}
            <div>
                <div>{"Liquids: "}</div>
                <LiquidEntriesGroupForNew initial={defaultLiquids} updateParent={setLiquids}/>
            </div>
            <div>
                <div>{"Nutrients (per Liter): "}</div>{/* TODO: per 400mL?*/}
                <NutrientsEntriesGroupForNew initial={defaultNutrients}
                                             updateParent={setNutrients}/>
            </div>
            <div>
                <div>{"Sugars (per Liter): "}</div>{/* TODO: per 400mL?*/}
                <SugarEntriesGroupForNew initial={defaultSugars} updateParent={setSugars}/>
            </div>
            <div>
                <div>{"Additives (per Liter): "}</div>{/* TODO: per 400mL?*/}
                <AdditiveEntriesGroupForNew initial={defaultAdditives} updateParent={setAdditives}/>
            </div>
            <div>
                <div>{"Antibiotics: "}</div>
                <AntibioticEntriesGroupForNew initial={defaultAntibiotics}
                                              updateParent={setAntibiotics}/>
            </div>
            <NewEntryNotes setNotes={setNotes}/>
            {/* SUBMIT AREA */}
            <CreateNewEntryButton onSubmit={newAgarRecipeSubmit}/>
        </NewEntryFormWrapper>
    )
}

export function agarPer400mL(agar: number) {
    return <div>{"(" + (agar * 2.0 / 5.0) + " g/400mL)"}</div>
}

// TODO: validate working! changed on 6/15/26
export const AgarRecipeArea = ({agarRecipeBinId,agarRecipe}: { agarRecipeBinId?: string, agarRecipe?:AgarRecipeData }) => {
    const [recipe, setRecipe] = useState<AgarRecipeData | undefined>(agarRecipe)
    const firstArea = ()=>{
        if (recipe || (!agarRecipe && !agarRecipeBinId)) {
            return <div>{"Agar Recipe: "}</div>
        }
        return <div>{"Agar Recipe ID: "}</div>
    }
    const linkArea = ()=>{
        if (recipe) {
            return <EntryLinkForId props={{
                displayId: recipe.name, // TODO: add id?
                linkId: recipe._id,
                entryType: "agarRecipe"
            }}/>
        }
        if (agarRecipeBinId) {
            return <>
                <EntryLinkForId props={{
                    displayId: agarRecipeBinId,
                    linkId: agarRecipeBinId,
                    entryType: "agarRecipe"
                }}/>
                <button className={"basicButtonSmall"} onClick={e=>{
                    e.stopPropagation()
                    // TODO: LOAD THE NAME! Validate works!
                    DoGetRequest("agarRecipe", agarRecipeBinId, AssertAgarRecipe, (e)=>{
                        console.error("failed to get agar recipe: "+JSON.stringify(e));
                    }).then(setRecipe)
                }}>{"Load Name"}</button>
            </>
        }
        return <div>{"unknown"}</div>
    }
    return <div className={"agarRecipeArea"}>
        <div>{firstArea()}</div>
        <div>{linkArea()}</div>
    </div>
}


export function AgarRecipeSelector(
    {
        doSelect,
        allowCreate
    }: {
        doSelect: (val: AgarRecipeData | undefined) => void,
        allowCreate?: boolean
    }) {
    const table = (items: AgarRecipeData[]): JSX.Element => {
        return <AgarRecipeSelectorTable data={items} onClick={doSelect}
                                        withLink={true}/>
    }

    return <ExistingDualSelector entryType={"agarRecipe"} entryTypes={"agarRecipes"} doSelect={doSelect}
                                 asserter={AssertAgarRecipe}
                                 table={table}>
        {allowCreate && <NewAgarRecipeForm handlers={{onCreate: doSelect, isTopLevel: false}}/>}
    </ExistingDualSelector>
}

export function AgarRecipeListPageTable({data, onClick, withLink}: ListPageItems<AgarRecipeData>) {
    let cols: ListTableColumn<AgarRecipeData>[] = [
        NewColumn("ID", (v) => v._id, true),
        NewColumn("Name", (v) => v.name), // TODO: shortname? fit?
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
        NewColumn("Antibiotics", (v) => {
            return <div>
                {v.antibiotics && v.antibiotics.map((v, i) => {
                    return <div key={i}>{v}</div>
                })}
            </div>
        }), // TODO: fit?
        NewColumn("Last Updated", (v) => {
            return NumberToDateStr(v.lastUpdated)
        }) // TODO: fit on last?

    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: AgarRecipeData) => {
            return <EntryLinkWrapper props={{entry: v, openInNewTab: true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable className={"text-xs"} cols={cols} data={data} onClick={onClick} newClass={v => {
        return new AgarRecipeData(v)
    }}/>
}

export function AgarRecipeSelectorTable({data, onClick, withLink}: ListPageItems<AgarRecipeData>) {
    let cols: ListTableColumn<AgarRecipeData>[] = [
        NewColumn("Name", (v) => v.name), // TODO: shortname?
        NewColumn("ID", (v) => v._id),
        NewColumn("Last Updated", (v) => {
            return NumberToDateStr(v.lastUpdated)
        })
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: AgarRecipeData) => {
            return <ViewInNewTabButton entry={v}/>
        })]
    }
    return <ListPageTable className={"text-xs"} cols={cols} data={data} onClick={onClick} newClass={v => {
        return new AgarRecipeData(v)
    }}/>
}
